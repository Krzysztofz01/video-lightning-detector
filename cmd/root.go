package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"

	"github.com/Krzysztofz01/video-lightning-detector/internal/options"
)

var (
	Version       string
	DeveloperMode string
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChannel
		cancel()
	}()

	rootCmd.SetArgs(args)
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stdout, "Failure: %s\n", err)
		os.Exit(1)
	}

	cancel()
	os.Exit(0)
}

var rootCmd = &cobra.Command{
	Use:  "vld",
	Long: "A video analysis tool that allows to detect and export frames that have captured lightning strikes.",
}
