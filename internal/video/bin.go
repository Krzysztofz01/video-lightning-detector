package video

import (
	"fmt"
	"os/exec"
	"strings"
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

func GetBinariesVersions() (map[string]string, error) {
	var (
		binaries []string          = []string{ffmpegBinaryName, ffprobeBinaryName}
		versions map[string]string = make(map[string]string, len(binaries))
	)

	for _, bin := range binaries {
		cmd := exec.Command(bin, "-version")

		stdout, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("video: unavailable binary %s due to version check error: %w", bin, err)
		}

		found := false
		for _, stdoutPart := range strings.Split(string(stdout), "\n") {
			if !strings.HasPrefix(stdoutPart, bin) {
				continue
			}

			versions[bin] = stdoutPart
			found = true
			break
		}

		if !found {
			return nil, fmt.Errorf("video: failed to access binary %s version value", bin)
		}
	}

	return versions, nil
}
