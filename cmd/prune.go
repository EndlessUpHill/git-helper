package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)



var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete merged branches",
	Long: `Clean up local branches that have been merged.

This command helps you:
1. Fetch and prune remote branches
2. Find local branches that are merged
3. Safely delete merged branches

Useful when:
- You have many stale branches
- Want to clean up after merging PRs
- Need to remove old feature branches

Example:
  githelper prune              # Interactive branch cleanup
  githelper prune --force      # Delete without confirmation
  githelper prune --main dev   # Use 'dev' as main branch`,
	RunE: runPrune,
}

func init() {
	rootCmd.AddCommand(pruneCmd)
	pruneCmd.Flags().StringVar(&mainBranch, "main", "main", "main branch name")
	pruneCmd.Flags().BoolVar(&force, "force", false, "delete without confirmation")
}

func runPrune(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Fetch and prune
	fmt.Println("🔄 Fetching and pruning remote branches...")
	fetchCmd := exec.Command("git", "fetch", "-p")
	fetchCmd.Stderr = os.Stderr
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch and prune: %w", err)
	}

	// Get merged branches
	branches, err := getMergedBranches()
	if err != nil {
		return err
	}

	if len(branches) == 0 {
		fmt.Println("✅ No merged branches to clean up!")
		return nil
	}

	// Show branches to delete
	fmt.Println("\nMerged branches to delete:")
	for _, branch := range branches {
		fmt.Printf("- %s\n", branch)
	}

	// Confirm deletion
	if !force {
		if !confirmAction() {
			fmt.Println("❌ Operation cancelled")
			return nil
		}
	}

	// Delete branches
	deleted := 0
	for _, branch := range branches {
		fmt.Printf("🗑️  Deleting branch '%s'...\n", branch)
		deleteCmd := exec.Command("git", "branch", "-d", branch)
		deleteCmd.Stderr = os.Stderr
		if err := deleteCmd.Run(); err != nil {
			fmt.Printf("⚠️  Failed to delete branch '%s': %v\n", branch, err)
			continue
		}
		deleted++
	}

	fmt.Printf("✅ Successfully deleted %d merged branch(es)!\n", deleted)
	return nil
}

func getMergedBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--merged", mainBranch)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list merged branches: %w", err)
	}

	var branches []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		// Skip current and main branches
		if branch != "" && !strings.HasPrefix(branch, "*") && branch != mainBranch {
			branches = append(branches, branch)
		}
	}

	return branches, nil
} 