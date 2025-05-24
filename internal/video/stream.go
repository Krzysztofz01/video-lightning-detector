package video

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

type VideoStream interface {
	GetInputDimensions() (int, int)
	GetOutputDimensions() (int, int)
	SetScale(s float64) error
	SetScaleAlgorithm(a options.ScaleAlgorithm) error
	SetBbox(x, y, w, h int) error
	SetFrameBuffer(buffer []byte) error
	SetHttpHeaders(headers http.Header) error
	Read() error
	Close()
}

type videoStream struct {
	Url            string
	HttpHeaders    http.Header
	Dim            utils.Vec2i
	BboxAnchor     utils.Vec2i
	BboxDim        utils.Vec2i
	Scale          float64
	ScaleAlgorithm options.ScaleAlgorithm
	Fps            float64
	FrameBuffer    []byte
	Process        *exec.Cmd
	Pipe           io.ReadCloser
}

func (v *videoStream) SetHttpHeaders(headers http.Header) error {
	if headers == nil {
		return fmt.Errorf("video: the provided http headers reference is nil")
	}

	v.HttpHeaders = headers
	return nil
}

func (v *videoStream) GetInputDimensions() (int, int) {
	return v.Dim.X, v.Dim.Y
}

func (v *videoStream) GetScaledDimensions() (int, int) {
	wF := float64(v.Dim.X) * v.Scale
	hF := float64(v.Dim.Y) * v.Scale

	return int(wF), int(hF)
}

func (v *videoStream) GetOutputBbox() (int, int, int, int) {
	xF := float64(v.BboxAnchor.X) * v.Scale
	yF := float64(v.BboxAnchor.Y) * v.Scale
	wF := float64(v.BboxDim.X) * v.Scale
	hF := float64(v.BboxDim.Y) * v.Scale

	return int(xF), int(yF), int(wF), int(hF)
}

func (v *videoStream) GetOutputDimensions() (int, int) {
	_, _, w, h := v.GetOutputBbox()

	return w, h
}

func (v *videoStream) SetScaleAlgorithm(a options.ScaleAlgorithm) error {
	if v.IsInitialized() {
		return fmt.Errorf("video: can not change scale algorithm after initialization")
	}

	if !options.IsValidScaleAlgorithm(a) {
		return fmt.Errorf("video: the specified scale algorithm is invalid")
	}

	v.ScaleAlgorithm = a
	return nil
}

func (v *videoStream) SetScale(s float64) error {
	if v.IsInitialized() {
		return fmt.Errorf("video: can not change scale after initialization")
	}

	if s <= 0 || s > 1 {
		return fmt.Errorf("video: provided scale factor is out of range")
	}

	v.Scale = s
	return nil
}

func (v *videoStream) SetBbox(x, y, w, h int) error {
	if v.IsInitialized() {
		return fmt.Errorf("video: can not change bbox after initialization")
	}

	if w <= 0 || h <= 0 {
		return fmt.Errorf("video: the video stream bbox sizes can not be negative or zero")
	}

	if x < 0 || x >= v.Dim.X || y < 0 || y >= v.Dim.Y {
		return fmt.Errorf("video: the video stream bbox anchor is out of the video range")
	}

	if x+w >= v.Dim.X {
		return fmt.Errorf("video: the video stream bbox horizontaly overflows the video bounds")
	}

	if y+h >= v.Dim.Y {
		return fmt.Errorf("video: the video stream bbox verticaly overflows the video bounds")
	}

	v.BboxAnchor = utils.Vec2i{X: x, Y: y}
	v.BboxDim = utils.Vec2i{X: w, Y: h}
	return nil
}

func (v *videoStream) IsBboxUsed() bool {
	return v.Dim.X != v.BboxDim.X || v.Dim.Y != v.BboxDim.Y
}

func (v *videoStream) SetFrameBuffer(buffer []byte) error {
	if v.IsInitialized() {
		return fmt.Errorf("video: can not change the frame buffer after initialization")
	}

	w, h := v.GetOutputDimensions()
	size := w * h * frameChannelDepth

	if len(buffer) != size {
		return fmt.Errorf("video: the target buffer size of %d does not match the required buffer length of %d", len(buffer), size)
	}

	v.FrameBuffer = buffer
	return nil
}

func (v *videoStream) Read() error {
	if !v.IsInitialized() {
		if err := v.Init(); err != nil {
			return fmt.Errorf("video: failed to initliaze frame reading video stream stream: %w", err)
		}
	}

	if _, err := io.ReadFull(v.Pipe, v.FrameBuffer); err == nil {
		return nil
	} else if err == io.EOF {
		return io.EOF
	} else if err == io.ErrUnexpectedEOF {
		return fmt.Errorf("video: failed to read the video frame data via the process pipe due to invalid data length")
	} else {
		return fmt.Errorf("video: failed to read the video frame data via the process pipe: %w", err)
	}
}

func (v *videoStream) Close() {
	defer func() {
		if err := recover(); err != nil {
			v.Process = nil
			v.Pipe = nil
			v.FrameBuffer = nil
		}
	}()

	if v.Process != nil {
		if v.Process.ProcessState == nil || !v.Process.ProcessState.Exited() {
			v.Process.Process.Kill()
		}

		v.Process.Wait()
		v.Process = nil
	}

	if v.Pipe != nil {
		v.Pipe.Close()
		v.Pipe = nil
	}

	if v.FrameBuffer != nil {
		v.FrameBuffer = nil
	}
}

func (v *videoStream) IsInitialized() bool {
	return v.Process != nil && v.Pipe != nil && v.FrameBuffer != nil
}

func (v *videoStream) Init() error {
	if v.IsInitialized() {
		return fmt.Errorf("video: video stream reading process can not be reinitialized")
	}

	var (
		filters []string = make([]string, 0, 16)
		args    []string = make([]string, 0, 32)
	)

	if v.Scale != 1 {
		w, h := v.GetScaledDimensions()

		var algorithm string
		switch v.ScaleAlgorithm {
		case options.Default:
		case options.Bilinear:
			algorithm = "bilinear"
		case options.Bicubic:
			algorithm = "bicubic"
		case options.NearestNeighbour:
			algorithm = "nearest"
		case options.Lanczos:
			algorithm = "lanczos"
		case options.Area:
			algorithm = "area"
		default:
			return fmt.Errorf("video: invalid corrupted scale algorithm value")
		}

		if v.ScaleAlgorithm == options.Default {
			filters = append(filters, fmt.Sprintf("scale=%d:%d", w, h))
		} else {
			filters = append(filters, fmt.Sprintf("scale=%d:%d:flags=%s", w, h, algorithm))
		}
	}

	if v.IsBboxUsed() {
		x, y, w, h := v.GetOutputBbox()
		filters = append(filters, fmt.Sprintf("crop=%d:%d:%d:%d", w, h, x, y))
	}

	args = append(args, "-i", v.Url)
	args = append(args, "-loglevel", "quiet")
	args = append(args, "-hide_banner")
	args = append(args, "-f", "image2pipe")
	args = append(args, "-pix_fmt", "rgba")
	args = append(args, "-vcodec", "rawvideo")
	args = append(args, "-map", "0:v:0")

	if len(v.HttpHeaders) > 0 {
		headerEntries := make([]string, 0, len(v.HttpHeaders))
		for key, values := range v.HttpHeaders {
			value := strings.Join(values, ", ")
			headerEntries = append(headerEntries, fmt.Sprintf("%s: %s", key, value))
		}

		// NOTE: According to FFmpeg documentation the headers should be separated with CRLF
		headers := strings.Join(headerEntries, string([]byte{0x0D, 0x0A}))

		args = append(args, "-headers", headers)
	}

	if len(filters) > 0 {
		filter := strings.Join(filters, ",")
		args = append(args, "-vf", filter)
	}

	args = append(args, "-")

	cmd := exec.Command(ffmpegBinaryName, args...)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("video: failed to access the video reading process pipe: %w", err)
	}

	if v.FrameBuffer == nil {
		w, h := v.GetOutputDimensions()
		size := w * h * frameChannelDepth

		v.FrameBuffer = make([]byte, size)
	}

	v.Process = cmd
	v.Pipe = pipe

	if err := v.Process.Start(); err != nil {
		return fmt.Errorf("video: failed to start the video reading proces: %w", err)
	}

	return nil
}

func NewVideoStream(url string) (VideoStream, error) {
	if !utils.IsValidUrl(url) {
		return nil, fmt.Errorf("video: the video stream url is invalid")
	}

	if ok, err := AreBinariesAvailable(); !ok && err != nil {
		return nil, fmt.Errorf("video: the required video processing binaries are not available: %w", err)
	}

	probe, err := probeVideoStream(url)
	if err != nil {
		return nil, fmt.Errorf("video: failed to probe the target video stream: %w", err)
	}

	return &videoStream{
		Url:            url,
		HttpHeaders:    http.Header{},
		Dim:            utils.Vec2i{X: probe.Width, Y: probe.Height},
		BboxAnchor:     utils.Vec2i{X: 0, Y: 0},
		BboxDim:        utils.Vec2i{X: probe.Width, Y: probe.Height},
		Scale:          1,
		ScaleAlgorithm: options.Default,
		Fps:            probe.Fps,
		FrameBuffer:    nil,
		Process:        nil,
		Pipe:           nil,
	}, nil
}
