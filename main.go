package main

import (
	"os"

	"github.com/Krzysztofz01/video-lightning-detector/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:]))
}
