package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	fixLineEndings bool
	cleanUntracked bool
)

var refreshCmd = &cobra.Command{
	Use:   "refresh [file...]",
	Short: "Fix Git index and line ending issues",
	Long: `Fix issues where Git shows files as modified when they haven't changed.

This command helps fix common Git index issues:
1. Refreshes Git's index
2. Optionally fixes line ending issues
3. Optionally removes untracked files

Common scenarios this fixes:
- Git shows files as modified but you haven't changed them
- Line ending (CRLF/LF) issues causing false modifications
- Need to clean up and start fresh

Example:
  githelper refresh              # Refresh all files
  githelper refresh file.txt     # Refresh specific file
  githelper refresh --crlf       # Fix line ending issues
  githelper refresh --clean      # Also remove untracked files`,
	RunE: runRefresh,
}

func init() {
	rootCmd.AddCommand(refreshCmd)
	refreshCmd.Flags().BoolVar(&fixLineEndings, "crlf", false, "fix line ending issues")
	refreshCmd.Flags().BoolVar(&cleanUntracked, "clean", false, "remove untracked files and directories")
}

func runRefresh(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Check for uncommitted changes that would be lost
	if cleanUntracked {
		statusCmd := exec.Command("git", "status", "--porcelain")
		status, err := statusCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to check git status: %w", err)
		}
		if len(status) > 0 {
			fmt.Println("⚠️  WARNING: This will remove all untracked files and directories!")
			if !confirmAction() {
				fmt.Println("❌ Operation cancelled")
				return nil
			}
		}
	}

	// Fix line endings if requested
	if fixLineEndings {
		fmt.Println("🔧 Fixing line endings...")
		if err := fixCRLFIssues(); err != nil {
			return err
		}
	}

	// Reset index for specified files or all files
	fmt.Println("🔄 Refreshing Git index...")
	checkoutArgs := []string{"checkout", "--"}
	if len(args) > 0 {
		checkoutArgs = append(checkoutArgs, args...)
	} else {
		checkoutArgs = append(checkoutArgs, ".")
	}

	checkoutCmd := exec.Command("git", checkoutArgs...)
	checkoutCmd.Stderr = os.Stderr
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to refresh index: %w", err)
	}

	// Clean untracked files if requested
	if cleanUntracked {
		fmt.Println("🧹 Removing untracked files...")
		cleanCmd := exec.Command("git", "clean", "-fd")
		cleanCmd.Stderr = os.Stderr
		if err := cleanCmd.Run(); err != nil {
			return fmt.Errorf("failed to clean untracked files: %w", err)
		}
	}

	// Reset to HEAD
	resetCmd := exec.Command("git", "reset", "--hard", "HEAD")
	resetCmd.Stderr = os.Stderr
	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to HEAD: %w", err)
	}

	fmt.Println("✅ Git index refreshed successfully!")
	return nil
}

func fixCRLFIssues() error {
	// Disable autocrlf
	configCmd := exec.Command("git", "config", "core.autocrlf", "false")
	if err := configCmd.Run(); err != nil {
		return fmt.Errorf("failed to configure line endings: %w", err)
	}

	// Re-normalize all files
	normalizeCmd := exec.Command("git", "add", "--renormalize", ".")
	if err := normalizeCmd.Run(); err != nil {
		return fmt.Errorf("failed to renormalize files: %w", err)
	}

	return nil
} 