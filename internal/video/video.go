package video

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

// FIXME: Add support for hardware accelerated decoding
// FIXME: Add support for exporting specific frames

const (
	depth int = 4
)

type vec2i struct {
	X, Y int
}

type Video interface {
	SetScale(s float64) error
	SetBbox(x, y, w, h int) error
	Width() int
	Height() int
	Frames() int
	SetFrameBuffer(buffer []byte) error
	Read() bool
	Close()
}

type video struct {
	FilePath    string
	Dim         vec2i
	BboxAnchor  vec2i
	BboxDim     vec2i
	Scale       float64
	Duration    float64
	Fps         float64
	FramesCount int
	FrameBuffer []byte
	Process     *exec.Cmd
	Pipe        io.ReadCloser
}

func (v *video) SetScale(s float64) error {
	if v.IsInitialized() {
		return fmt.Errorf("video: can not change scale after initialization")
	}

	if s <= 0 || s > 1 {
		return fmt.Errorf("video: provided scale factor is out of range")
	}

	v.Scale = s
	return nil
}

func (v *video) SetBbox(x, y, w, h int) error {
	if v.IsInitialized() {
		return fmt.Errorf("video: can not change bbox after initialization")
	}

	if x+w >= v.Dim.X {
		return fmt.Errorf("video: the video bbox horizontaly overflows the video bounds")
	}

	if y+h >= v.Dim.Y {
		return fmt.Errorf("video: the video bbox verticaly overflows the video bounds")
	}

	v.BboxAnchor = vec2i{X: x, Y: y}
	v.BboxDim = vec2i{X: w, Y: h}
	return nil
}

func (v *video) Frames() int {
	return v.FramesCount
}

func (v *video) Width() int {
	return int(float64(v.BboxDim.X) * v.Scale)
}

func (v *video) Height() int {
	return int(float64(v.BboxDim.Y) * v.Scale)
}

func (v *video) IsBboxUsed() bool {
	return v.Dim.X != v.BboxDim.X || v.Dim.Y != v.BboxDim.Y
}

func (v *video) SetFrameBuffer(buffer []byte) error {
	if v.IsInitialized() {
		return fmt.Errorf("video: can not change the frame buffer after initialization")
	}

	size := v.Width() * v.Height() * depth
	if len(buffer) != size {
		return fmt.Errorf("video: the target buffer size of %d does not match the required buffer length of %d", len(buffer), size)
	}

	v.FrameBuffer = buffer
	return nil
}

// TODO: Better handling of errors and EOF
func (v *video) Read() bool {
	if !v.IsInitialized() {
		if err := v.Init(); err != nil {
			return false
		}
	}

	_, err := io.ReadFull(v.Pipe, v.FrameBuffer)
	if err != nil {
		v.Close()
		return false
	}

	// TODO: Remove debug logs
	// fmt.Printf("Ex: %d\nAc: %d\n", v.Width()*v.Height()*depth, read)

	return true
}

func (v *video) Close() {
	if v.Pipe != nil {
		v.Pipe.Close()
		v.Pipe = nil
	}

	if v.Process != nil {
		v.Process.Process.Kill()
		v.Process = nil
	}

	if v.FrameBuffer != nil {
		v.FrameBuffer = nil
	}
}

func (v *video) IsInitialized() bool {
	return v.Process != nil && v.Pipe != nil && v.FrameBuffer != nil
}

func (v *video) Init() error {
	if v.IsInitialized() {
		return fmt.Errorf("video: video reading process can not be reinitialized")
	}

	args := []string{
		"-i", v.FilePath,
		"-loglevel", "quiet",
		"-hide_banner",
		"-f", "image2pipe",
		"-pix_fmt", "rgba",
		"-vcodec", "rawvideo",
		// "-hwaccel", "auto",
		"-map", "0:v:0",
	}

	filterExpressions := make([]string, 0, 8)

	if v.Scale != 1 {
		expression := fmt.Sprintf("scale=iw*%[1]f:ih*%[1]f", v.Scale)
		filterExpressions = append(filterExpressions, expression)
	}

	if v.IsBboxUsed() {
		x := int(float64(v.BboxAnchor.X) * v.Scale)
		y := int(float64(v.BboxAnchor.Y) * v.Scale)
		w := int(float64(v.BboxDim.X) * v.Scale)
		h := int(float64(v.BboxDim.Y) * v.Scale)

		expression := fmt.Sprintf("crop=%d:%d:%d:%d", w, h, x, y)
		filterExpressions = append(filterExpressions, expression)
	}

	if len(filterExpressions) > 0 {
		filterExpression := strings.Join(filterExpressions, ",")
		args = append(args, "-vf", filterExpression)
	}

	args = append(args, "-")

	// TODO: remove debug logs
	fmt.Printf("\n%s\n\n", strings.Join(args, " "))
	cmd := exec.Command(ffmpegBinaryName, args...)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("video: failed to access the video reading process pipe: %w", err)
	}

	if v.FrameBuffer == nil {
		v.FrameBuffer = make([]byte, v.BboxAnchor.X*v.BboxAnchor.Y*depth)
	}

	v.Process = cmd
	v.Pipe = pipe

	if err := v.Process.Start(); err != nil {
		return fmt.Errorf("video: failed to start the video reading proces: %w", err)
	}

	return nil
}

func NewVideo(path string) (Video, error) {
	if !utils.FileExists(path) {
		return nil, fmt.Errorf("video: the video file specified by the does not exist")
	}

	if ok, err := AreBinariesAvailable(); !ok && err != nil {
		return nil, fmt.Errorf("video: the required video processing binaries are not available: %w", err)
	}

	probe, err := probeVideo(path)
	if err != nil {
		return nil, fmt.Errorf("video: failed to probe the target video file: %w", err)
	}

	return &video{
		FilePath:    path,
		Dim:         vec2i{X: probe.Width, Y: probe.Height},
		BboxAnchor:  vec2i{X: 0, Y: 0},
		BboxDim:     vec2i{X: probe.Width, Y: probe.Height},
		Scale:       1,
		Duration:    probe.Duration,
		Fps:         probe.Fps,
		FramesCount: probe.Frames,
		FrameBuffer: nil,
		Process:     nil,
		Pipe:        nil,
	}, nil
}
