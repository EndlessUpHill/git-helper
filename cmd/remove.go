package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EndlessUphill/git-helper/internal/git"
	"github.com/spf13/cobra"
)

var (
	removeForce bool
)

// removeFileCmd represents the remove command
var removeFileCmd = &cobra.Command{
	Use:   "remove [file]",
	Short: "Remove a file from git history",
	Long: `Remove a file from git history. This command rewrites git history to remove a file
from all commits. This is a destructive operation that should be used with extreme caution.

WARNING: This command rewrites git history and should NEVER be used on:
- Shared repositories where others have cloned or forked your work
- Repositories where others have based work on the commits you're modifying
- Public repositories where others might have referenced specific commits

This command is intended for:
- Personal repositories where you're the only contributor
- Removing sensitive files that were accidentally committed
- Cleaning up large files that were accidentally committed
- Fixing mistakes in your local repository before pushing

After running this command, you will need to force push your changes:
git push --force

Example:
  git-helper remove path/to/sensitive/file.txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath)
		}

		// Convert to absolute path
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Get relative path from git root
		gitRoot, err := git.GetGitRoot()
		if err != nil {
			return fmt.Errorf("failed to get git root: %w", err)
		}

		relPath, err := filepath.Rel(gitRoot, absPath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Confirm with user
		fmt.Printf("WARNING: This will permanently remove '%s' from git history.\n", relPath)
		fmt.Println("This operation cannot be undone and will rewrite git history.")
		fmt.Println("Make sure you understand the implications before proceeding.")
		fmt.Println("\nIf you're sure you want to proceed, run:")
		fmt.Printf("git filter-branch --force --index-filter \"git rm --cached --ignore-unmatch %s\" --prune-empty --tag-name-filter cat -- --all\n", relPath)
		fmt.Println("\nAfter the operation completes, run:")
		fmt.Println("rm -rf .git/refs/original/")
		fmt.Println("git reflog expire --expire=now --all")
		fmt.Println("git gc --prune=now --aggressive")
		fmt.Println("\nFinally, force push your changes:")
		fmt.Println("git push --force")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeFileCmd)
	removeFileCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Skip confirmation and execute immediately")
} 