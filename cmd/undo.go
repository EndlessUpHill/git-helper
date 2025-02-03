package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	hardReset bool
	numCommits int
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last push to remote",
	Long: `Undo the last push to remote while optionally keeping changes locally.
	
Two modes available:
- Soft (default): Keeps changes staged in your working directory
- Hard: Completely removes the changes

Example: githelper undo        # soft reset of last commit
         githelper undo --hard # hard reset of last commit
         githelper undo -n 3   # undo last 3 commits`,
	RunE: runUndo,
}

func init() {
	rootCmd.AddCommand(undoCmd)
	flags := undoCmd.Flags()
	flags.BoolVar(&hardReset, "hard", false, "completely remove changes (hard reset)")
	flags.IntVarP(&numCommits, "num", "n", 1, "number of commits to undo")
}

func runUndo(cmd *cobra.Command, args []string) error {
	// Check if current directory is a git repository
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Confirm with user before proceeding
	if !confirmUndo() {
		fmt.Println("❌ Undo operation cancelled")
		return nil
	}

	// Determine reset type
	resetType := "--soft"
	if hardReset {
		resetType = "--hard"
	}

	// Reset local commits
	resetCmd := exec.Command("git", "reset", resetType, fmt.Sprintf("HEAD~%d", numCommits))
	resetCmd.Stdout = os.Stdout
	resetCmd.Stderr = os.Stderr
	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("failed to reset commits: %w", err)
	}

	// Force push to remote
	pushCmd := exec.Command("git", "push", "origin", "HEAD", "--force-with-lease")
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to force push: %w", err)
	}

	// Print success message
	if hardReset {
		fmt.Printf("✅ Successfully removed last %d commit(s) and pushed changes\n", numCommits)
	} else {
		fmt.Printf("✅ Successfully undid last %d commit(s) while keeping changes locally\n", numCommits)
	}

	return nil
}

func confirmUndo() bool {
	fmt.Printf("⚠️  Warning: This will undo the last %d commit(s) ", numCommits)
	if hardReset {
		fmt.Print("and remove all changes")
	} else {
		fmt.Print("but keep changes locally")
	}
	fmt.Print("\nAre you sure you want to continue? [y/N]: ")

	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
} 