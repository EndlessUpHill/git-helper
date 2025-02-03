package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	depth       int
	singleBranch bool
	noTags      bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone <repository> [directory]",
	Short: "Clone repositories with size optimization",
	Long: `Optimized git clone with size reduction options.

This command helps you clone repositories efficiently:
1. Shallow clone with specified depth
2. Single branch clone option
3. Option to skip downloading tags

Useful for:
- Huge repositories that take too long to clone
- When you only need recent history
- CI/CD environments where full history isn't needed

Example:
  githelper clone https://github.com/org/repo.git        # Normal clone
  githelper clone --depth 1 https://github.com/org/repo  # Shallow clone
  githelper clone --single-branch org/repo               # Clone only default branch`,
	Args: cobra.MinimumNArgs(1),
	RunE: runClone,
}

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().IntVarP(&depth, "depth", "d", 0, "create a shallow clone with specified depth")
	cloneCmd.Flags().BoolVar(&singleBranch, "single-branch", false, "clone only the default branch")
	cloneCmd.Flags().BoolVar(&noTags, "no-tags", false, "don't clone any tags")
}

func runClone(cmd *cobra.Command, args []string) error {
	repo := args[0]
	
	// Handle directory argument
	var directory string
	if len(args) > 1 {
		directory = args[1]
	} else {
		// Extract directory name from repo URL
		directory = getDefaultDirectory(repo)
	}

	// Normalize repository URL
	repo = normalizeRepoURL(repo)

	// Build clone command with options
	cloneArgs := []string{"clone"}

	if depth > 0 {
		cloneArgs = append(cloneArgs, "--depth", fmt.Sprintf("%d", depth))
	}

	if singleBranch {
		cloneArgs = append(cloneArgs, "--single-branch")
	}

	if noTags {
		cloneArgs = append(cloneArgs, "--no-tags")
	}

	// Add progress display
	cloneArgs = append(cloneArgs, "--progress")

	// Add repository URL and directory
	cloneArgs = append(cloneArgs, repo, directory)

	// Show what we're doing
	fmt.Printf("🔄 Cloning repository: %s\n", repo)
	if depth > 0 {
		fmt.Printf("📏 Shallow clone with depth: %d\n", depth)
	}
	if singleBranch {
		fmt.Println("🌿 Cloning only the default branch")
	}
	if noTags {
		fmt.Println("🏷️  Skipping tag download")
	}

	// Run the clone command
	cloneCmd := exec.Command("git", cloneArgs...)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr

	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Get repo size after cloning
	size, err := getRepoSize(directory)
	if err == nil {
		fmt.Printf("📦 Repository size: %s\n", formatSize(size))
	}

	fmt.Printf("✅ Repository cloned successfully to: %s\n", directory)
	return nil
}

func normalizeRepoURL(repo string) string {
	// Handle GitHub shorthand (org/repo)
	if !strings.Contains(repo, "://") && !strings.Contains(repo, "@") {
		if !strings.HasSuffix(repo, ".git") {
			repo += ".git"
		}
		return "https://github.com/" + repo
	}
	return repo
}

func getDefaultDirectory(repo string) string {
	// Remove .git suffix if present
	repo = strings.TrimSuffix(repo, ".git")
	
	// Handle different URL formats
	if strings.Contains(repo, "://") {
		parts := strings.Split(repo, "/")
		return parts[len(parts)-1]
	}
	
	// Handle GitHub shorthand
	parts := strings.Split(repo, "/")
	return parts[len(parts)-1]
}

func getRepoSize(directory string) (int64, error) {
	var size int64
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}