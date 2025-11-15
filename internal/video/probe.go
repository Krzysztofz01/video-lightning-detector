package video

import (
	"bufio"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Krzysztofz01/video-lightning-detector/internal/utils"
)

const frameChannelDepth int = 4

type streamStrategy int

const (
	singleStreamStrategy streamStrategy = iota
	firstStreamStrategy
)

type probeToken string

const (
	tokenStreamOpen   probeToken = "[STREAM]"
	tokenStreamClose  probeToken = "[/STREAM]"
	tokenFormatOpen   probeToken = "[FORMAT]"
	tokenFormatClose  probeToken = "[/FORMAT]"
	tokenProgramOpen  probeToken = "[PROGRAM]"
	tokenProgramClose probeToken = "[/PROGRAM]"
	notApplicable     probeToken = "N/A"
)

type VideoProbe struct {
	Width    int
	Height   int
	Duration float64
	Frames   int
	Fps      float64
}

func probeVideoFile(path string) (VideoProbe, error) {
	if !utils.FileExists(path) {
		return VideoProbe{}, fmt.Errorf("video: the probe target video file does not exist")
	}

	return probeVideo(path, singleStreamStrategy)
}

func probeVideoStream(url string) (VideoProbe, error) {
	if !utils.IsValidUrl(url) {
		return VideoProbe{}, fmt.Errorf("video: the probe target video stream url format is invalid")
	}

	return probeVideo(url, firstStreamStrategy)
}

func probeVideo(resource string, strategy streamStrategy) (VideoProbe, error) {
	cmd := exec.Command(
		ffprobeBinaryName,
		"-select_streams", "v",
		"-loglevel", "quiet",
		"-show_entries", "stream=width,height,rotation,nb_frames,r_frame_rate",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=0:nokey=0",
		resource,
	)

	stdout, err := cmd.Output()
	if err != nil {
		return VideoProbe{}, fmt.Errorf("video: failed to probe the video: %w", err)
	}

	result, err := parseProbeResult(stdout, strategy)
	if err != nil {
		return VideoProbe{}, fmt.Errorf("video: failed to parse the video probe result: %w", err)
	}

	return result, nil
}

func parseProbeResult(stdout []byte, strategy streamStrategy) (VideoProbe, error) {
	reader := strings.NewReader(string(stdout))
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	var (
		probe              VideoProbe = VideoProbe{}
		stateRotate        bool       = false
		stateStreamCount   int        = 0
		stateStreamParity  int        = 0
		stateFormatParity  int        = 0
		stateProgramParity int        = 0
	)

	var (
		key   string
		value string
	)

scanning:
	for scanner.Scan() {
		switch scanner.Text() {
		case string(tokenStreamOpen):
			{
				switch strategy {
				case singleStreamStrategy:
					{
						if stateStreamCount > 0 {
							return VideoProbe{}, fmt.Errorf("video: video files containing multiple video streams are not supported")
						}
					}
				case firstStreamStrategy:
					{
						if stateStreamCount == 1 && stateStreamParity == 0 {
							break scanning
						}
					}
				}

				stateStreamCount += 1
				stateStreamParity += 1
				continue
			}
		case string(tokenStreamClose):
			{
				stateStreamParity -= 1
				continue
			}
		case string(tokenFormatOpen):
			{
				stateFormatParity += 1
				continue
			}
		case string(tokenFormatClose):
			{
				stateFormatParity -= 1
				continue
			}
		case string(tokenProgramOpen):
			{
				stateProgramParity += 1
				continue
			}
		case string(tokenProgramClose):
			{
				stateProgramParity -= 1
				continue
			}
		default:
		}

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
				if value == string(notApplicable) {
					probe.Duration = 0
					continue
				}

				if v, err := strconv.ParseFloat(value, 64); err != nil {
					return VideoProbe{}, fmt.Errorf("video: failed to parse the probed video duration: %w", err)
				} else {
					probe.Duration = v
				}
			}
		case "nb_frames":
			{
				if value == string(notApplicable) {
					probe.Frames = 0
					continue
				}

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
					stateRotate = true
				}
			}
		default:
			{
			}
		}
	}

	if stateStreamCount != 1 {
		return VideoProbe{}, fmt.Errorf("video: the video contains a invalid number of video streams")
	}

	if stateStreamParity != 0 || stateFormatParity != 0 || stateProgramParity != 0 {
		return VideoProbe{}, fmt.Errorf("video: failed to probe the video due to invalid probe result format")
	}

	if stateRotate {
		probe.Width, probe.Height = probe.Height, probe.Width
	}

	// NOTE: Additional calculation due to the fact that nb_frames is metadata based and is not exact. The greater count value is preferred for now.
	calculatedFramesCount := int(math.Ceil(probe.Fps * probe.Duration))
	if probe.Frames < calculatedFramesCount {
		probe.Frames = calculatedFramesCount
	}

	return probe, nil
}
