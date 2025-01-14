package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "githelper",
	Short: "A CLI tool to simplify complex GitHub workflows",
	Long: `GitHelper is a command-line tool that simplifies complex GitHub workflows
that are not straightforward with basic Git commands. It provides various
utilities to manage repositories, branches, and common Git operations.`,
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug logging")
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.githelper.yaml)")
}