package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var rescueCmd = &cobra.Command{
	Use:   "rescue [branch-name]",
	Short: "Rescue commits from detached HEAD state",
	Long: `Create a new branch from detached HEAD state.

This command helps when you're stuck in a "detached HEAD" state:
1. Verifies if you're in detached HEAD
2. Shows recent commits for reference
3. Creates a new branch from current position

Useful when:
- You checked out a specific commit without -b
- You're in "detached HEAD" state
- You need to save your work before switching branches

Example:
  githelper rescue              # Interactive branch creation
  githelper rescue new-branch   # Create specific branch name`,
	RunE: runRescue,
}

func init() {
	rootCmd.AddCommand(rescueCmd)
}

func runRescue(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Check if we're in detached HEAD state
	isDetached, err := isDetachedHead()
	if err != nil {
		return err
	}
	if !isDetached {
		return fmt.Errorf("not in detached HEAD state. This command is only needed when HEAD is detached")
	}

	// Show current position
	fmt.Println("🔍 Current HEAD position:")
	showCmd := exec.Command("git", "log", "--oneline", "-n", "1")
	showCmd.Stdout = os.Stdout
	showCmd.Stderr = os.Stderr
	if err := showCmd.Run(); err != nil {
		return fmt.Errorf("failed to show current commit: %w", err)
	}

	// Show recent commits
	fmt.Println("\n📜 Recent commits:")
	logCmd := exec.Command("git", "log", "--oneline", "-n", "5")
	logCmd.Stdout = os.Stdout
	logCmd.Stderr = os.Stderr
	if err := logCmd.Run(); err != nil {
		return fmt.Errorf("failed to show recent commits: %w", err)
	}

	// Get branch name
	var branchName string
	if len(args) > 0 {
		branchName = args[0]
	} else {
		branchName = getBranchNameInteractive()
		if branchName == "" {
			return fmt.Errorf("no branch name provided")
		}
	}

	// Create new branch
	fmt.Printf("\n🌱 Creating new branch '%s' from current position...\n", branchName)
	checkoutCmd := exec.Command("git", "checkout", "-b", branchName)
	checkoutCmd.Stderr = os.Stderr
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	fmt.Printf("✅ Successfully created branch '%s'!\n", branchName)
	fmt.Println("\nYou can now continue working on this branch.")
	return nil
}

func isDetachedHead() (bool, error) {
	// Get current HEAD reference
	refCmd := exec.Command("git", "symbolic-ref", "-q", "HEAD")
	err := refCmd.Run()
	
	// If the command fails, we're in detached HEAD
	if err != nil {
		// Verify it's actually a detached HEAD and not some other error
		headCmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
		if headCmd.Run() == nil {
			return true, nil
		}
		return false, fmt.Errorf("failed to check HEAD state: %w", err)
	}
	
	return false, nil
}

func getBranchNameInteractive() string {
	// Get current commit message for suggestion
	msgCmd := exec.Command("git", "log", "-1", "--pretty=%B")
	msg, err := msgCmd.Output()
	if err != nil {
		msg = []byte("")
	}

	// Generate suggestion from commit message
	suggestion := generateBranchName(string(msg))

	fmt.Printf("\nSuggested branch name: %s\n", suggestion)
	fmt.Print("Enter branch name (or press Enter to use suggestion): ")
	
	var input string
	fmt.Scanln(&input)
	
	if input == "" {
		return suggestion
	}
	return input
}

func generateBranchName(commitMsg string) string {
	// Clean up commit message
	msg := strings.TrimSpace(commitMsg)
	msg = strings.Split(msg, "\n")[0] // First line only
	
	// Remove common prefixes
	prefixes := []string{"feat:", "fix:", "chore:", "docs:", "style:", "refactor:", "test:"}
	for _, prefix := range prefixes {
		msg = strings.TrimPrefix(msg, prefix)
	}
	
	// Clean up and format
	msg = strings.TrimSpace(msg)
	msg = strings.ToLower(msg)
	msg = strings.ReplaceAll(msg, " ", "-")
	
	// Remove special characters
	msg = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, msg)
	
	// Limit length
	if len(msg) > 30 {
		msg = msg[:30]
	}
	
	// Ensure it starts with a letter
	if len(msg) == 0 || (msg[0] >= '0' && msg[0] <= '9') {
		msg = "branch-" + msg
	}
	
	return msg
} 