package cmd

import (
	"fmt"

	"github.com/EndlessUphill/git-helper/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GitHelper %s\n", version.Version)
		fmt.Printf("Commit: %s\n", version.CommitHash)
		fmt.Printf("Built: %s\n", version.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
} 