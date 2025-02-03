package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var cherryPickCmd = &cobra.Command{
	Use:   "cherry-pick <pr-number>",
	Short: "Cherry-pick commits from a PR",
	Long: `Interactively cherry-pick commits from a pull request.

This command helps you:
1. Fetch the PR as a local branch
2. Show commits in the PR
3. Let you select which commits to cherry-pick

Useful when:
- You want specific changes from a PR
- Need to apply fixes to multiple branches
- Want to test specific commits

Example:
  githelper cherry-pick 123     # Cherry-pick from PR #123`,
	Args: cobra.ExactArgs(1),
	RunE: runCherryPick,
}

func init() {
	rootCmd.AddCommand(cherryPickCmd)
}

func runCherryPick(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Parse PR number
	prNum, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid PR number: %s", args[0])
	}

	// Fetch PR
	fmt.Printf("🔄 Fetching PR #%d...\n", prNum)
	fetchCmd := exec.Command("git", "fetch", "origin", fmt.Sprintf("pull/%d/head:pr-%d", prNum, prNum))
	fetchCmd.Stderr = os.Stderr
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch PR: %w", err)
	}

	// Get commits using fzf
	commits, err := selectCommitsWithFzf(prNum)
	if err != nil {
		return err
	}
	if len(commits) == 0 {
		return fmt.Errorf("no commits selected")
	}

	// Cherry-pick each commit
	for _, commit := range commits {
		fmt.Printf("🍒 Cherry-picking commit %s...\n", commit[:8])
		cherryCmd := exec.Command("git", "cherry-pick", commit)
		cherryCmd.Stdout = os.Stdout
		cherryCmd.Stderr = os.Stderr
		if err := cherryCmd.Run(); err != nil {
			return fmt.Errorf("failed to cherry-pick commit %s: %w", commit[:8], err)
		}
	}

	fmt.Printf("✅ Successfully cherry-picked %d commit(s)!\n", len(commits))
	return nil
}

func selectCommitsWithFzf(prNum int) ([]string, error) {
	if !noFzf {
		if _, err := exec.LookPath("fzf"); err == nil {
			return selectCommitsWithFzfInteractive(prNum)
		}
	}
	return selectCommitsWithList(prNum)
}

func selectCommitsWithFzfInteractive(prNum int) ([]string, error) {
	// Get commit log
	logCmd := exec.Command("git", "log", "--oneline", "--reverse", fmt.Sprintf("pr-%d", prNum))
	output, err := logCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	// Create preview command
	previewCmd := "git show --color=always {1}"

	// Run fzf
	fzfCmd := exec.Command("fzf",
		"--multi",
		"--height", "50%",
		"--reverse",
		"--preview", previewCmd,
		"--preview-window", "right:50%")

	fzfCmd.Stdin = strings.NewReader(string(output))
	fzfCmd.Stderr = os.Stderr

	fzfOutput, err := fzfCmd.Output()
	if err != nil {
		return nil, nil // User cancelled
	}

	// Extract commit hashes
	var commits []string
	lines := strings.Split(strings.TrimSpace(string(fzfOutput)), "\n")
	for _, line := range lines {
		commits = append(commits, strings.Fields(line)[0])
	}

	return commits, nil
}

func selectCommitsWithList(prNum int) ([]string, error) {
	// Show commits
	fmt.Printf("\nCommits in PR #%d:\n", prNum)
	logCmd := exec.Command("git", "log", "--oneline", "--reverse", fmt.Sprintf("pr-%d", prNum))
	logCmd.Stdout = os.Stdout
	logCmd.Stderr = os.Stderr
	if err := logCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to show commits: %w", err)
	}

	// Get commit hashes
	fmt.Print("\nEnter commit hashes to cherry-pick (space-separated): ")
	var input string
	fmt.Scanln(&input)

	if input == "" {
		return nil, nil
	}

	return strings.Fields(input), nil
} 