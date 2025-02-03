package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	upstreamURL string
)

var syncForkCmd = &cobra.Command{
	Use:   "sync-fork",
	Short: "Sync fork with upstream repository",
	Long: `Synchronize your fork with the upstream repository.

This command helps you keep your fork up to date by:
1. Setting up upstream remote if needed
2. Fetching upstream changes
3. Rebasing your changes on top of upstream
4. Safely pushing to your fork

Useful when:
- Maintaining a fork of another repository
- Need to get latest changes from upstream
- Want to keep your fork in sync

Example:
  githelper sync-fork                              # Sync with detected upstream
  githelper sync-fork --upstream user/repo         # Sync with specific upstream
  githelper sync-fork --branch develop            # Sync specific branch`,
	RunE: runSyncFork,
}

func init() {
	rootCmd.AddCommand(syncForkCmd)
	syncForkCmd.Flags().StringVar(&upstreamURL, "upstream", "", "upstream repository URL or path (user/repo)")
	syncForkCmd.Flags().StringVar(&mainBranch, "branch", "main", "main branch name (main or master)")
}

func runSyncFork(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Check for uncommitted changes
	if hasChanges, err := hasUncommittedChanges(); err != nil {
		return err
	} else if hasChanges {
		return fmt.Errorf("you have uncommitted changes. Please commit or stash them first")
	}

	// Setup upstream if needed
	if err := setupUpstream(); err != nil {
		return err
	}

	// Fetch upstream
	fmt.Println("🔄 Fetching upstream changes...")
	fetchCmd := exec.Command("git", "fetch", "upstream")
	fetchCmd.Stderr = os.Stderr
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch upstream: %w", err)
	}

	// Get current branch
	currentBranch, err := getCurrentBranch()
	if err != nil {
		return err
	}

	// Rebase on upstream
	fmt.Printf("📥 Rebasing on upstream/%s...\n", mainBranch)
	rebaseCmd := exec.Command("git", "rebase", fmt.Sprintf("upstream/%s", mainBranch))
	rebaseCmd.Stdout = os.Stdout
	rebaseCmd.Stderr = os.Stderr
	if err := rebaseCmd.Run(); err != nil {
		fmt.Println("\n⚠️  Rebase failed. Please resolve conflicts and run:")
		fmt.Println("git rebase --continue")
		fmt.Println("Then run this command again")
		return fmt.Errorf("rebase failed: %w", err)
	}

	// Push to origin
	fmt.Printf("📤 Pushing to origin/%s...\n", currentBranch)
	pushCmd := exec.Command("git", "push", "origin", currentBranch, "--force-with-lease")
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push to origin: %w", err)
	}

	fmt.Printf("✅ Successfully synced fork with upstream/%s!\n", mainBranch)
	return nil
}

func setupUpstream() error {
	// Check if upstream remote exists
	remoteCmd := exec.Command("git", "remote")
	output, err := remoteCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list remotes: %w", err)
	}

	remotes := strings.Fields(string(output))
	hasUpstream := false
	for _, remote := range remotes {
		if remote == "upstream" {
			hasUpstream = true
			break
		}
	}

	if !hasUpstream {
		if upstreamURL == "" {
			// Try to detect upstream from origin URL
			originURL, err := getOriginURL()
			if err != nil {
				return fmt.Errorf("upstream not configured and could not detect: %w", err)
			}

			upstreamURL = detectUpstreamURL(originURL)
			if upstreamURL == "" {
				return fmt.Errorf("could not detect upstream repository. Please specify with --upstream")
			}
		}

		// Add upstream remote
		fmt.Printf("🔗 Adding upstream remote: %s\n", upstreamURL)
		addCmd := exec.Command("git", "remote", "add", "upstream", upstreamURL)
		addCmd.Stderr = os.Stderr
		if err := addCmd.Run(); err != nil {
			return fmt.Errorf("failed to add upstream remote: %w", err)
		}
	}

	return nil
}

func getOriginURL() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func detectUpstreamURL(originURL string) string {
	// Handle SSH format: git@github.com:user/repo.git
	if strings.HasPrefix(originURL, "git@") {
		parts := strings.Split(originURL, ":")
		if len(parts) != 2 {
			return ""
		}
		repoPath := strings.TrimSuffix(parts[1], ".git")
		if strings.Count(repoPath, "/") != 1 {
			return ""
		}
		return fmt.Sprintf("https://github.com/%s.git", repoPath)
	}

	// Handle HTTPS format: https://github.com/user/repo.git
	if strings.HasPrefix(originURL, "https://") {
		parts := strings.Split(originURL, "/")
		if len(parts) < 5 {
			return ""
		}
		// Remove fork's username from path
		parts = append(parts[:len(parts)-2], parts[len(parts)-1])
		return strings.Join(parts, "/")
	}

	return ""
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
} 