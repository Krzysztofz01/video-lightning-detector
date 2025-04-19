package video

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

// FIXME: Support for handling file with multiple streams (Check for [STREAM][/STREAM])
// FIXME: Support for handling file without streams (same as above)

type VideoProbe struct {
	Width    int
	Height   int
	Duration float64
	Frames   int
	Fps      float64
}

func probeVideo(path string) (VideoProbe, error) {
	if !utils.FileExists(path) {
		return VideoProbe{}, fmt.Errorf("video: the probe target video file does not exist")
	}

	cmd := exec.Command(
		ffprobeBinaryName,
		"-select_streams", "v",
		"-loglevel", "quiet",
		"-show_entries", "stream=width,height,rotation,nb_frames,r_frame_rate",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=0",
		path,
	)

	stdout, err := cmd.Output()
	if err != nil {
		return VideoProbe{}, fmt.Errorf("video: failed to probe the video: %w", err)
	}

	result, err := parseProbeResult(stdout)
	if err != nil {
		return VideoProbe{}, fmt.Errorf("video: failed to parse the video probe result: %w", err)
	}

	return result, nil
}

func parseProbeResult(stdout []byte) (VideoProbe, error) {
	reader := strings.NewReader(string(stdout))
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	var (
		probe  VideoProbe = VideoProbe{}
		rotate bool       = false
		key    string
		value  string
	)

	for scanner.Scan() {
		if textParts := strings.Split(scanner.Text(), "="); len(textParts) != 2 {
			return VideoProbe{}, fmt.Errorf("video: failed to parse the probe result due to invalid result format")
		} else {
			key = textParts[0]
			value = textParts[1]
		}

		switch key {
		case "width":
			{
				if v, err := strconv.ParseInt(value, 10, 0); err != nil {
					return VideoProbe{}, fmt.Errorf("video: failed to parse the probed video width: %w", err)
				} else {
					probe.Width = int(v)
				}
			}
		case "height":
			{
				if v, err := strconv.ParseInt(value, 10, 0); err != nil {
					return VideoProbe{}, fmt.Errorf("video: failed to parse the probed video height: %w", err)
				} else {
					probe.Height = int(v)
				}
			}
		case "duration":
			{
				if v, err := strconv.ParseFloat(value, 64); err != nil {
					return VideoProbe{}, fmt.Errorf("video: failed to parse the probed video duration: %w", err)
				} else {
					probe.Duration = v
				}
			}
		case "nb_frames":
			{
				if v, err := strconv.ParseInt(value, 10, 0); err != nil {
					return VideoProbe{}, fmt.Errorf("video: failed to parse the probed video frames count: %w", err)
				} else {
					probe.Frames = int(v)
				}
			}
		case "r_frame_rate":
			{
				fpsParts := strings.Split(value, "/")
				if len(fpsParts) != 2 {
					return VideoProbe{}, fmt.Errorf("video: failed to parse the probed video frame rate due to invalid format")
				}

				var (
					frames  float64
					seconds float64
					err     error
				)

				if frames, err = strconv.ParseFloat(fpsParts[0], 64); err != nil {
					return VideoProbe{}, fmt.Errorf("video: failed to parse the probed video frame rate seconds value: %w", err)
				}

				if seconds, err = strconv.ParseFloat(fpsParts[1], 64); err != nil {
					return VideoProbe{}, fmt.Errorf("video: failed to prase the probed video frame rate frames value: %w", err)
				}

				probe.Fps = frames / seconds
			}
		case "tag:rotate":
			{
				if value == "90" || value == "270" {
					rotate = true
				}
			}
		default:
			{
			}
		}
	}

	if rotate {
		probe.Width, probe.Height = probe.Height, probe.Width
	}

	return probe, nil
}
