package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a deleted branch",
	Long: `Restore a deleted branch using git reflog.

This command helps you recover a branch that was deleted before merging.
It will:
1. Show you a list of recent git actions
2. Let you select the commit to restore
3. Create a new branch from that commit

Example: githelper restore`,
	RunE: runRestore,
}

var noFzf bool

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().BoolVar(&noFzf, "no-fzf", false, "disable fzf usage even if available")
}

func runRestore(cmd *cobra.Command, args []string) error {
	// Check if current directory is a git repository
	if err := checkGitRepo(); err != nil {
		return err
	}

	fmt.Println("🔍 Searching for git history...")

	// Get git reflog
	reflogCmd := exec.Command("git", "reflog")
	reflogOutput, err := reflogCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git reflog: %w", err)
	}

	// Parse and display reflog entries
	entries := parseReflog(string(reflogOutput))
	if len(entries) == 0 {
		return fmt.Errorf("no git history found")
	}

	// Let user select a commit
	commit := selectCommit(entries)
	if commit == "" {
		fmt.Println("❌ No commit selected")
		return nil
	}

	// Get branch name from user
	branchName := getBranchName()
	if branchName == "" {
		fmt.Println("❌ No branch name provided")
		return nil
	}

	// Create new branch
	checkoutCmd := exec.Command("git", "checkout", "-b", branchName, commit)
	checkoutCmd.Stdout = os.Stdout
	checkoutCmd.Stderr = os.Stderr
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	fmt.Printf("✅ Branch '%s' restored successfully!\n", branchName)
	return nil
}

func parseReflog(reflog string) []ReflogEntry {
	var entries []ReflogEntry
	scanner := bufio.NewScanner(strings.NewReader(reflog))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			entries = append(entries, ReflogEntry{
				Hash:        parts[0],
				Description: parts[1],
			})
		}
	}
	return entries
}

func selectCommit(entries []ReflogEntry) string {
	if !noFzf {
		if _, err := exec.LookPath("fzf"); err == nil {
			return selectCommitWithFzf(entries)
		}
	}
	return selectCommitWithList(entries)
}

func selectCommitWithFzf(entries []ReflogEntry) string {
	// Create input for fzf
	var input strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(&input, "%s %s\n", entry.Hash[:8], entry.Description)
	}

	// Create fzf command
	fzfCmd := exec.Command("fzf", "--ansi", "--height", "50%", "--reverse")
	fzfCmd.Stderr = os.Stderr
	fzfCmd.Stdin = strings.NewReader(input.String())

	// Get fzf output
	output, err := fzfCmd.Output()
	if err != nil {
		return "" // User cancelled or error occurred
	}

	// Parse the selected line to get the commit hash
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return ""
	}

	// Extract full hash from the selected entry
	parts := strings.Fields(selected)
	if len(parts) == 0 {
		return ""
	}
	selectedShortHash := parts[0]

	// Find the matching entry with full hash
	for _, entry := range entries {
		if strings.HasPrefix(entry.Hash, selectedShortHash) {
			return entry.Hash
		}
	}

	return ""
}

func selectCommitWithList(entries []ReflogEntry) string {
	fmt.Println("\nRecent git actions:")
	for i, entry := range entries {
		if i >= 20 { // Show only last 20 entries
			break
		}
		fmt.Printf("%2d: %s - %s\n", i+1, entry.Hash[:8], entry.Description)
	}

	fmt.Print("\nSelect commit number (or press Enter to cancel): ")
	var input string
	fmt.Scanln(&input)

	if input == "" {
		return ""
	}

	var index int
	if _, err := fmt.Sscanf(input, "%d", &index); err != nil || index < 1 || index > len(entries) {
		return ""
	}

	return entries[index-1].Hash
}

func getBranchName() string {
	fmt.Print("Enter a name for the restored branch: ")
	var branchName string
	fmt.Scanln(&branchName)
	return strings.TrimSpace(branchName)
} 