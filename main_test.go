package main

import (
	"os"
	"strings"
	"testing"

	"github.com/Krzysztofz01/video-lightning-detector/cmd"
)

func BenchmarkVideoLightningDetectorFromEnvArgs(b *testing.B) {
	args := os.Getenv("VLD_CLI_ARGS")

	cmd.Execute(strings.Split(args, " "))
}
