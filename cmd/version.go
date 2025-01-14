package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	CommitHash = "none"
	BuildDate  = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GitHelper %s\n", Version)
		fmt.Printf("Commit: %s\n", CommitHash)
		fmt.Printf("Built: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
} 