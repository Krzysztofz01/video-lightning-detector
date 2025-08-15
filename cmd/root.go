package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
)

var (
	LogLevel options.LogLevel
)

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().VarP(&LogLevel, "log-level", "l", "The verbosity of the log messages printed to the standard output.")
}

func Execute(args []string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stdout, "Unexpected failure: %s\n", err)
			os.Exit(1)
		}
	}()

	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stdout, "Failure: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

var rootCmd = &cobra.Command{
	Use:  "vld",
	Long: "A video analysis tool that allows to detect and export frames that have captured lightning strikes.",
}
