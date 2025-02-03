package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	forcePush bool
)

var purgeCmd = &cobra.Command{
	Use:   "purge [file]",
	Short: "Remove sensitive files from git history",
	Long: `Completely remove a file from git history.

This command helps you remove sensitive files (like API keys) from your git history.
It will:
1. Let you select a file to remove
2. Remove all traces of the file from git history
3. Optionally force push the changes

⚠️  WARNING: This rewrites git history! Use with caution, especially on shared repositories.

Example:
  githelper purge                  # Interactive file selection
  githelper purge config.json      # Remove specific file
  githelper purge --force-push     # Also force push changes`,
	RunE: runPurge,
}

func init() {
	rootCmd.AddCommand(purgeCmd)
	purgeCmd.Flags().BoolVar(&forcePush, "force-push", false, "force push changes after purging")
}

func runPurge(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	var fileToPurge string
	var err error

	if len(args) > 0 {
		fileToPurge = args[0]
	} else {
		// Interactive file selection
		fileToPurge, err = selectFile()
		if err != nil {
			return err
		}
		if fileToPurge == "" {
			return fmt.Errorf("no file selected")
		}
	}

	// Confirm action
	fmt.Printf("\n⚠️  WARNING: This will permanently remove '%s' from git history!\n", fileToPurge)
	fmt.Println("This action CANNOT be undone and will rewrite git history.")
	if !confirmAction() {
		fmt.Println("❌ Operation cancelled")
		return nil
	}

	// Remove file from git history
	fmt.Printf("\n🚨 Removing '%s' from git history...\n", fileToPurge)
	filterCmd := exec.Command("git", "filter-branch", "--force",
		"--index-filter", fmt.Sprintf("git rm --cached --ignore-unmatch %s", fileToPurge),
		"--prune-empty", "--tag-name-filter", "cat", "--", "--all")
	
	filterCmd.Stdout = os.Stdout
	filterCmd.Stderr = os.Stderr
	
	if err := filterCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove file from history: %w", err)
	}

	// Force push if requested
	if forcePush {
		fmt.Println("\n🔄 Force pushing changes...")
		pushCmd := exec.Command("git", "push", "origin", "--force", "--all")
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		if err := pushCmd.Run(); err != nil {
			return fmt.Errorf("failed to force push: %w", err)
		}
	} else {
		fmt.Println("\n⚠️  Changes are local only. To push them:")
		fmt.Println("git push origin --force --all")
	}

	fmt.Println("✅ File removed from git history!")
	return nil
}

func selectFile() (string, error) {
	// Try using fzf if available
	if !noFzf {
		if _, err := exec.LookPath("fzf"); err == nil {
			return selectFileWithFzf()
		}
	}
	return selectFileWithList()
}

func selectFileWithFzf() (string, error) {
	// Get list of files
	lsCmd := exec.Command("git", "ls-files")
	lsOutput, err := lsCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	// Check if bat is available for preview
	batAvailable := false
	if _, err := exec.LookPath("bat"); err == nil {
		batAvailable = true
	}

	// Create fzf command with preview
	previewCmd := "cat {}"
	if batAvailable {
		previewCmd = "bat --style=numbers --color=always {}"
	}

	fzfCmd := exec.Command("fzf",
		"--height", "50%",
		"--reverse",
		"--preview", previewCmd,
		"--preview-window", "right:50%")
	
	fzfCmd.Stdin = strings.NewReader(string(lsOutput))
	fzfCmd.Stderr = os.Stderr

	// Get fzf output
	output, err := fzfCmd.Output()
	if err != nil {
		return "", nil // User cancelled
	}

	return strings.TrimSpace(string(output)), nil
}

func selectFileWithList() (string, error) {
	// Get list of files
	lsCmd := exec.Command("git", "ls-files")
	output, err := lsCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}

	// Display files
	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	fmt.Println("\nTracked files:")
	for i, file := range files {
		fmt.Printf("%2d: %s\n", i+1, file)
	}

	// Get user selection
	fmt.Print("\nSelect file number (or press Enter to cancel): ")
	var input string
	fmt.Scanln(&input)

	if input == "" {
		return "", nil
	}

	var index int
	if _, err := fmt.Sscanf(input, "%d", &index); err != nil || index < 1 || index > len(files) {
		return "", fmt.Errorf("invalid selection")
	}

	return files[index-1], nil
}

func confirmAction() bool {
	fmt.Print("Are you sure you want to continue? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
} 