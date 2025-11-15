package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Krzysztofz01/video-lightning-detector/internal/video"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version numbers.",
	Long:  "Print the version numbers of VLD and dependency binaries.",
	Run: func(cmd *cobra.Command, args []string) {
		version := strings.Builder{}
		version.WriteString(Version)

		if DeveloperMode == "true" {
			version.WriteString(" [Developer Mode]")
		}

		fmt.Fprintf(os.Stdout, "video-lightning-detector version %s Copyright (c) Krzysztof Zoń\n", version.String())

		versions, err := video.GetBinariesVersions()
		if err != nil {
			fmt.Fprintf(os.Stdout, "Required FFmpeg binaries are not available!\n")
		}

		for _, version := range versions {
			fmt.Fprintf(os.Stdout, "%s\n", version)
		}
	},
}
