package video

import (
	"fmt"
	"os/exec"
)

const (
	ffmpegBinaryName  string = "ffmpeg"
	ffprobeBinaryName string = "ffprobe"
)

func AreBinariesAvailable() (bool, error) {
	binaries := []string{ffmpegBinaryName, ffprobeBinaryName}
	for _, bin := range binaries {
		cmd := exec.Command(bin, "-version")

		if err := cmd.Run(); err != nil {
			return false, fmt.Errorf("video: unavailable binary %s due to check error: %w", bin, err)
		}
	}

	return true, nil
}
