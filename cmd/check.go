package cmd

import (
	"fmt"
	"os"

	"github.com/Krzysztofz01/video-lightning-detector/internal/video"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if the environment is correctly configured.",
	Long:  "Check if the required dependency binaries are available in the current environment.",
	Run: func(cmd *cobra.Command, args []string) {
		if ok, _ := video.AreBinariesAvailable(); !ok {
			fmt.Fprintf(os.Stdout, "The required FFmpeg binaries are not available!\n")
			return
		}

		fmt.Fprintf(os.Stdout, "The environment is correctly configured for VLD usage.\n")
	},
}
