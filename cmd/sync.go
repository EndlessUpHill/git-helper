package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	noStash bool
)

var syncCmd = &cobra.Command{
	Use:   "sync [branch]",
	Short: "Safely sync local and remote changes",
	Long: `Safely synchronize local commits with remote changes.

This command helps when your push is rejected due to remote changes:
1. Stashes your working changes (optional)
2. Pulls remote changes with rebase
3. Restores your working changes

Useful when:
- Push is rejected due to remote changes
- You want to update local branch without merge commits
- You have local changes you don't want to lose

Example:
  githelper sync              # Sync current branch
  githelper sync main        # Sync specific branch
  githelper sync --no-stash  # Skip stashing (if working tree is clean)
  githelper sync --force     # Force sync even with uncommitted changes`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&noStash, "no-stash", false, "skip stashing changes")
	syncCmd.Flags().BoolVar(&force, "force", false, "force sync even with uncommitted changes")
}

func runSync(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Get current branch if not specified
	branch := "HEAD"
	if len(args) > 0 {
		branch = args[0]
	}

	// Check for uncommitted changes
	hasChanges, err := hasUncommittedChanges()
	if err != nil {
		return err
	}

	if hasChanges && !noStash {
		// Stash changes if needed
		fmt.Println("📦 Stashing local changes...")
		if err := stashChanges(); err != nil {
			return err
		}
		defer func() {
			if err := popStash(); err != nil {
				fmt.Printf("⚠️  Failed to restore stashed changes: %v\n", err)
				fmt.Println("Your changes are still in the stash. Use 'git stash pop' to restore them.")
			}
		}()
	} else if hasChanges {
		if !force {
			return fmt.Errorf("you have uncommitted changes. Use --force to proceed anyway, or commit/stash your changes")
		}
		fmt.Println("⚠️  Proceeding with uncommitted changes (forced)")
	}

	// Fetch remote changes
	fmt.Println("🔄 Fetching remote changes...")
	fetchCmd := exec.Command("git", "fetch", "origin")
	fetchCmd.Stderr = os.Stderr
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch remote changes: %w", err)
	}

	// Pull with rebase
	fmt.Println("📥 Pulling remote changes with rebase...")
	pullCmd := exec.Command("git", "pull", "--rebase", "origin", branch)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		if hasChanges && !noStash {
			fmt.Println("\n⚠️  Rebase failed. Your original changes are safe in the stash.")
			fmt.Println("Resolve the conflicts and run 'git stash pop' to restore your changes.")
		}
		return fmt.Errorf("failed to pull with rebase: %w", err)
	}

	fmt.Println("✅ Successfully synchronized with remote!")
	return nil
}

func hasUncommittedChanges() (bool, error) {
	statusCmd := exec.Command("git", "status", "--porcelain")
	output, err := statusCmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	return len(output) > 0, nil
}

func stashChanges() error {
	stashCmd := exec.Command("git", "stash", "save", "--include-untracked", 
		fmt.Sprintf("Automatic stash by githelper sync at %s", getCurrentTimestamp()))
	stashCmd.Stderr = os.Stderr
	return stashCmd.Run()
}

func popStash() error {
	fmt.Println("📦 Restoring your local changes...")
	popCmd := exec.Command("git", "stash", "pop")
	popCmd.Stdout = os.Stdout
	popCmd.Stderr = os.Stderr
	return popCmd.Run()
}

func getCurrentTimestamp() string {
	output, err := exec.Command("date", "+%Y-%m-%d %H:%M:%S").Output()
	if err != nil {
		return "unknown"
	}
	return string(output)
} 
