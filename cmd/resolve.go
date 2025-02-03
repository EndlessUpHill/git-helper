package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve [file]",
	Short: "Resolve merge conflicts easily",
	Long: `Resolve git merge conflicts by choosing between "ours" or "theirs".

This command helps you resolve merge conflicts quickly when you didn't edit the file:
1. Lists all files with conflicts
2. Let's you choose which file to resolve
3. Allows you to pick between your version (ours) or their version (theirs)
4. Stages the resolved file

Example:
  githelper resolve              # Interactive file selection
  githelper resolve config.json  # Resolve specific file`,
	RunE: runResolve,
}

func init() {
	rootCmd.AddCommand(resolveCmd)
}

func runResolve(cmd *cobra.Command, args []string) error {
	// Check if there are any conflicts
	if !hasConflicts() {
		return fmt.Errorf("no merge conflicts found")
	}

	var fileToResolve string
	var err error

	if len(args) > 0 {
		// Verify the specified file has conflicts
		fileToResolve = args[0]
		if !isFileConflicted(fileToResolve) {
			return fmt.Errorf("specified file '%s' has no conflicts", fileToResolve)
		}
	} else {
		// Interactive file selection
		fileToResolve, err = selectConflictedFile()
		if err != nil {
			return err
		}
		if fileToResolve == "" {
			return fmt.Errorf("no file selected")
		}
	}

	// Show diff and get resolution choice
	if err := showConflictDiff(fileToResolve); err != nil {
		fmt.Println("⚠️  Failed to show diff, continuing anyway...")
	}

	choice := getResolutionChoice(fileToResolve)
	
	// Resolve the conflict
	var checkoutFlag string
	switch choice {
	case "o", "ours":
		checkoutFlag = "--ours"
	case "t", "theirs":
		checkoutFlag = "--theirs"
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}

	// Checkout the chosen version
	checkoutCmd := exec.Command("git", "checkout", checkoutFlag, fileToResolve)
	checkoutCmd.Stderr = os.Stderr
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout version: %w", err)
	}

	// Stage the resolved file
	addCmd := exec.Command("git", "add", fileToResolve)
	addCmd.Stderr = os.Stderr
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage resolved file: %w", err)
	}

	fmt.Printf("✅ Conflict in '%s' resolved!\n", fileToResolve)
	return nil
}

func hasConflicts() bool {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.Output()
	return err == nil && len(output) > 0
}

func isFileConflicted(file string) bool {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, f := range files {
		if f == file {
			return true
		}
	}
	return false
}

func selectConflictedFile() (string, error) {
	// Try using fzf if available
	if !noFzf {
		if _, err := exec.LookPath("fzf"); err == nil {
			return selectConflictedFileWithFzf()
		}
	}
	return selectConflictedFileWithList()
}

func selectConflictedFileWithFzf() (string, error) {
	// Get list of conflicted files
	diffCmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list conflicted files: %w", err)
	}

	// Check if bat is available for preview
	batAvailable := false
	if _, err := exec.LookPath("bat"); err == nil {
		batAvailable = true
	}

	// Create preview command that shows the conflict markers
	previewCmd := "git diff {}"
	if batAvailable {
		previewCmd = "git diff {} | bat --style=numbers --color=always --language=diff"
	}

	fzfCmd := exec.Command("fzf",
		"--height", "50%",
		"--reverse",
		"--preview", previewCmd,
		"--preview-window", "right:60%")
	
	fzfCmd.Stdin = strings.NewReader(string(diffOutput))
	fzfCmd.Stderr = os.Stderr

	output, err := fzfCmd.Output()
	if err != nil {
		return "", nil // User cancelled
	}

	return strings.TrimSpace(string(output)), nil
}

func selectConflictedFileWithList() (string, error) {
	// Get list of conflicted files
	diffCmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	output, err := diffCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list conflicted files: %w", err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	fmt.Println("\nConflicted files:")
	for i, file := range files {
		fmt.Printf("%2d: %s\n", i+1, file)
	}

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

func showConflictDiff(file string) error {
	diffCmd := exec.Command("git", "diff", file)
	
	// Use bat if available
	if _, err := exec.LookPath("bat"); err == nil {
		diffCmd = exec.Command("sh", "-c", fmt.Sprintf("git diff %s | bat --style=numbers --color=always --language=diff", file))
	}
	
	diffCmd.Stdout = os.Stdout
	diffCmd.Stderr = os.Stderr
	return diffCmd.Run()
}

func getResolutionChoice(file string) string {
	fmt.Printf("\nResolving conflicts in '%s'\n", file)
	fmt.Println("Choose resolution:")
	fmt.Println("  (o)urs   - Keep our version (current branch)")
	fmt.Println("  (t)heirs - Keep their version (merging branch)")
	
	fmt.Print("\nYour choice [o/t]: ")
	var choice string
	fmt.Scanln(&choice)
	return strings.ToLower(choice)
} 