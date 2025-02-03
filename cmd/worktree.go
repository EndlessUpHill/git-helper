package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Manage git worktrees",
	Long: `Simplify git worktree management with easy-to-use commands.

This command helps you manage multiple git worktrees:
- Switch between worktrees
- Create new worktrees
- Remove worktrees
- Clean up merged worktrees
- Pull updates in worktrees

Example:
  githelper worktree switch     # Switch to another worktree
  githelper worktree create dev # Create new worktree for 'dev' branch
  githelper worktree cleanup    # Remove worktrees for merged branches`,
}

// Subcommands
var (
	switchCmd = &cobra.Command{
		Use:   "switch",
		Short: "Switch to another worktree",
		RunE:  runWorktreeSwitch,
	}

	createCmd = &cobra.Command{
		Use:   "create [branch]",
		Short: "Create a new worktree",
		Args:  cobra.ExactArgs(1),
		RunE:  runWorktreeCreate,
	}

	removeCmd = &cobra.Command{
		Use:   "remove [worktree]",
		Short: "Remove a worktree",
		Args:  cobra.ExactArgs(1),
		RunE:  runWorktreeRemove,
	}

	cleanupCmd = &cobra.Command{
		Use:   "cleanup",
		Short: "Remove worktrees for merged branches",
		RunE:  runWorktreeCleanup,
	}

	pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "Pull updates in a worktree",
		RunE:  runWorktreePull,
	}
)

func init() {
	rootCmd.AddCommand(worktreeCmd)
	worktreeCmd.AddCommand(switchCmd)
	worktreeCmd.AddCommand(createCmd)
	worktreeCmd.AddCommand(removeCmd)
	worktreeCmd.AddCommand(cleanupCmd)
	worktreeCmd.AddCommand(pullCmd)
}

func runWorktreeSwitch(cmd *cobra.Command, args []string) error {
	worktree, err := selectWorktree()
	if err != nil {
		return err
	}
	if worktree == "" {
		return fmt.Errorf("no worktree selected")
	}

	fmt.Printf("🔄 Switching to worktree: %s\n", worktree)
	if err := os.Chdir(worktree); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Print the new working directory
	pwd, err := os.Getwd()
	if err == nil {
		fmt.Printf("✅ Now in: %s\n", pwd)
	}
	return nil
}

func runWorktreeCreate(cmd *cobra.Command, args []string) error {
	branch := args[0]
	worktreePath := filepath.Join("..", branch)

	fmt.Printf("🌱 Creating worktree for branch '%s'...\n", branch)
	createCmd := exec.Command("git", "worktree", "add", worktreePath, branch)
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree: %w", err)
	}

	// Change to the new worktree
	if err := os.Chdir(worktreePath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	fmt.Printf("✅ Worktree created and switched to: %s\n", worktreePath)
	return nil
}

func runWorktreeRemove(cmd *cobra.Command, args []string) error {
	worktree := args[0]

	fmt.Printf("🗑️  Removing worktree: %s\n", worktree)
	removeCmd := exec.Command("git", "worktree", "remove", worktree)
	removeCmd.Stderr = os.Stderr
	if err := removeCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	fmt.Printf("✅ Worktree removed: %s\n", worktree)
	return nil
}

func runWorktreeCleanup(cmd *cobra.Command, args []string) error {
	// Get merged branches
	mergedCmd := exec.Command("git", "branch", "--merged", "main")
	mergedOutput, err := mergedCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get merged branches: %w", err)
	}

	branches := strings.Split(strings.TrimSpace(string(mergedOutput)), "\n")
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "main" || branch == "*" || branch == "" {
			continue
		}

		worktreePath := filepath.Join("..", branch)
		if _, err := os.Stat(worktreePath); err == nil {
			fmt.Printf("🗑️  Removing worktree for merged branch: %s\n", branch)
			removeCmd := exec.Command("git", "worktree", "remove", worktreePath)
			removeCmd.Run() // Ignore errors for cleanup
		}
	}

	fmt.Println("✅ Cleanup complete!")
	return nil
}

func runWorktreePull(cmd *cobra.Command, args []string) error {
	worktree, err := selectWorktree()
	if err != nil {
		return err
	}
	if worktree == "" {
		return fmt.Errorf("no worktree selected")
	}

	fmt.Printf("🔄 Pulling updates in worktree: %s\n", worktree)
	
	// Change to the selected worktree
	if err := os.Chdir(worktree); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Pull updates
	pullCmd := exec.Command("git", "pull")
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stderr
	if err := pullCmd.Run(); err != nil {
		return fmt.Errorf("failed to pull updates: %w", err)
	}

	fmt.Println("✅ Updates pulled successfully!")
	return nil
}

func selectWorktree() (string, error) {
	// Get worktree list
	listCmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := listCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []string
	lines := strings.Split(string(output), "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			worktrees = append(worktrees, path)
		}
	}

	if len(worktrees) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}

	// Try using fzf if available
	if !noFzf {
		if _, err := exec.LookPath("fzf"); err == nil {
			return selectWorktreeWithFzf(worktrees)
		}
	}
	return selectWorktreeWithList(worktrees)
}

func selectWorktreeWithFzf(worktrees []string) (string, error) {
	// Create input for fzf
	var input strings.Builder
	for _, worktree := range worktrees {
		fmt.Fprintln(&input, worktree)
	}

	// Create preview command that shows git status
	previewCmd := "cd {} && git status"

	fzfCmd := exec.Command("fzf",
		"--height", "50%",
		"--reverse",
		"--preview", previewCmd,
		"--preview-window", "right:50%")
	
	fzfCmd.Stdin = strings.NewReader(input.String())
	fzfCmd.Stderr = os.Stderr

	output, err := fzfCmd.Output()
	if err != nil {
		return "", nil // User cancelled
	}

	return strings.TrimSpace(string(output)), nil
}

func selectWorktreeWithList(worktrees []string) (string, error) {
	fmt.Println("\nAvailable worktrees:")
	for i, worktree := range worktrees {
		fmt.Printf("%2d: %s\n", i+1, worktree)
	}

	fmt.Print("\nSelect worktree number (or press Enter to cancel): ")
	var input string
	fmt.Scanln(&input)

	if input == "" {
		return "", nil
	}

	var index int
	if _, err := fmt.Sscanf(input, "%d", &index); err != nil || index < 1 || index > len(worktrees) {
		return "", fmt.Errorf("invalid selection")
	}

	return worktrees[index-1], nil
} 