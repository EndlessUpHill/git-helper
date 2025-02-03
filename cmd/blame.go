package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var blameCmd = &cobra.Command{
	Use:   "blame <file> <line>",
	Short: "Find line author across all commits",
	Long: `Track changes to a specific line across all commits.

This command helps you:
1. Show all changes to a specific line
2. See who modified it and why
3. Track the line's evolution

Useful when:
- Investigating code history
- Finding out who wrote specific code
- Understanding why code changed

Example:
  githelper blame main.go 42    # Show history of line 42 in main.go`,
	Args: cobra.ExactArgs(2),
	RunE: runBlame,
}

func init() {
	rootCmd.AddCommand(blameCmd)
}

func runBlame(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	file := args[0]
	line, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid line number: %s", args[1])
	}

	// Check if file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", file)
	}

	// Show line history
	fmt.Printf("📜 History for %s line %d:\n\n", file, line)
	logCmd := exec.Command("git", "log", "-L", fmt.Sprintf("%d,%d:%s", line, line, file))
	logCmd.Stdout = os.Stdout
	logCmd.Stderr = os.Stderr
	if err := logCmd.Run(); err != nil {
		return fmt.Errorf("failed to get line history: %w", err)
	}

	return nil
} 