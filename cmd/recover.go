package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover lost commits after a hard reset",
	Long: `Find and restore commits that were lost after a hard reset.

This command helps you recover lost work by:
1. Showing the git reflog with all recent actions
2. Letting you select a commit to restore to
3. Resetting your branch back to that commit

⚠️  WARNING: This will reset your current branch! Make sure to commit or stash changes.

Example:
  githelper recover    # Interactive commit selection`,
	RunE: runRecover,
}

func init() {
	rootCmd.AddCommand(recoverCmd)
}



func runRecover(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Check for uncommitted changes
	statusCmd := exec.Command("git", "status", "--porcelain")
	status, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}
	if len(status) > 0 {
		return fmt.Errorf("you have uncommitted changes. Please commit or stash them first")
	}

	fmt.Println("🔍 Searching for lost commits...")
	commit, err := selectCommitFromReflog()
	if err != nil {
		return err
	}
	if commit == "" {
		return fmt.Errorf("no commit selected")
	}

	// Confirm action
	fmt.Printf("\n⚠️  WARNING: This will reset your branch to commit: %s\n", commit)
	fmt.Println("This action will modify your current branch!")
	if !confirmAction() {
		fmt.Println("❌ Operation cancelled")
		return nil
	}

	// Reset to selected commit
	fmt.Printf("\n⏪ Resetting to commit: %s\n", commit)
	resetCmd := exec.Command("git", "reset", "--hard", commit)
	resetCmd.Stdout = os.Stdout
	resetCmd.Stderr = os.Stderr
	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to commit: %w", err)
	}

	fmt.Println("✅ Successfully reset to selected commit!")
	return nil
}

func selectCommitFromReflog() (string, error) {
	// Try using fzf if available
	if !noFzf {
		if _, err := exec.LookPath("fzf"); err == nil {
			return selectCommitWithFzfFromReflog()
		}
	}
	return selectCommitWithListFromReflog()
}

func getReflogEntries() ([]ReflogEntry, error) {
	reflogCmd := exec.Command("git", "reflog", "--pretty=%H %gd %gs")
	output, err := reflogCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get reflog: %w", err)
	}

	var entries []ReflogEntry
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 3)
		if len(parts) >= 3 {
			entries = append(entries, ReflogEntry{
				Hash:        parts[0],
				Action:      parts[1],
				Description: parts[2],
			})
		}
	}

	return entries, nil
}

func selectCommitWithFzfFromReflog() (string, error) {
	entries, err := getReflogEntries()
	if err != nil {
		return "", err
	}

	// Create input for fzf
	var input strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(&input, "%s %s: %s\n", 
			entry.Hash[:8], 
			entry.Action,
			entry.Description)
	}

	// Create preview command that shows commit details
	previewCmd := "git show --color=always {1}"

	fzfCmd := exec.Command("fzf",
		"--height", "50%",
		"--reverse",
		"--preview", previewCmd,
		"--preview-window", "right:50%",
		"--ansi")
	
	fzfCmd.Stdin = strings.NewReader(input.String())
	fzfCmd.Stderr = os.Stderr

	output, err := fzfCmd.Output()
	if err != nil {
		return "", nil // User cancelled
	}

	// Extract commit hash from selection
	selected := strings.TrimSpace(string(output))
	return strings.Fields(selected)[0], nil
}

func selectCommitWithListFromReflog() (string, error) {
	entries, err := getReflogEntries()
	if err != nil {
		return "", err
	}

	fmt.Println("\nRecent git actions:")
	for i, entry := range entries {
		if i >= 20 { // Show only last 20 entries
			break
		}
		fmt.Printf("%2d: %s %s: %s\n", 
			i+1,
			entry.Hash[:8],
			entry.Action,
			entry.Description)
	}

	fmt.Print("\nSelect action number (or press Enter to cancel): ")
	var input string
	fmt.Scanln(&input)

	if input == "" {
		return "", nil
	}

	var index int
	if _, err := fmt.Sscanf(input, "%d", &index); err != nil || index < 1 || index > len(entries) {
		return "", fmt.Errorf("invalid selection")
	}

	return entries[index-1].Hash, nil
} 