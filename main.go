package main

import (
	"os"

	"github.com/Krzysztofz01/video-lightning-detector/cmd"
)

func main() {
	cmd.Execute(os.Args[1:])
}
