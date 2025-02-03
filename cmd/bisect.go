package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var bisectCmd = &cobra.Command{
	Use:   "bisect",
	Short: "Find the commit that introduced a bug using git bisect",
	Long: `Interactive git bisect helper to find problematic commits.

This command helps you find which commit introduced a bug by using git's bisect feature.
It will guide you through the process:

1. Start by selecting a known GOOD commit (where everything worked)
2. Then select a known BAD commit (where the bug exists)
3. Git will checkout commits in between, and you test each one
4. For each commit, you tell git if it's good or bad
5. Git will narrow down the problematic commit using binary search

Example workflow:
  $ githelper bisect
  1. Select a known good commit (older version where bug didn't exist)
  2. Select a known bad commit (newer version where bug exists)
  3. Test each commit git checks out:
     - If the bug exists: run 'git bisect bad'
     - If the bug is gone: run 'git bisect good'
  4. Git will eventually find the exact commit that introduced the bug

Tips:
  - You can use 'git bisect reset' to abort the process
  - Write a test script to automate the verification
  - Use 'git bisect run ./test.sh' to automate the entire process`,
	RunE: runBisect,
}

func init() {
	rootCmd.AddCommand(bisectCmd)
}

func runBisect(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Start bisect
	fmt.Println("🔎 Starting Git Bisect...")
	if err := exec.Command("git", "bisect", "start").Run(); err != nil {
		return fmt.Errorf("failed to start git bisect: %w", err)
	}

	// Get good commit
	fmt.Println("\n📌 Select a known GOOD commit (where everything worked):")
	goodCommit, err := selectCommitForBisect()
	if err != nil {
		return fmt.Errorf("failed to select good commit: %w", err)
	}
	if goodCommit == "" {
		return fmt.Errorf("no good commit selected")
	}

	// Get bad commit
	fmt.Println("\n📌 Select a known BAD commit (where the bug exists):")
	badCommit, err := selectCommitForBisect()
	if err != nil {
		return fmt.Errorf("failed to select bad commit: %w", err)
	}
	if badCommit == "" {
		return fmt.Errorf("no bad commit selected")
	}

	// Mark good and bad commits
	if err := exec.Command("git", "bisect", "good", goodCommit).Run(); err != nil {
		return fmt.Errorf("failed to mark good commit: %w", err)
	}
	if err := exec.Command("git", "bisect", "bad", badCommit).Run(); err != nil {
		return fmt.Errorf("failed to mark bad commit: %w", err)
	}

	// Print instructions
	fmt.Println("\n🛠️  Git bisect is now running!")
	fmt.Println("\nInstructions:")
	fmt.Println("1. Git will checkout different commits for you to test")
	fmt.Println("2. Test if the bug exists in each commit")
	fmt.Println("3. Mark each commit using:")
	fmt.Println("   - git bisect good  (if the bug is NOT present)")
	fmt.Println("   - git bisect bad   (if the bug IS present)")
	fmt.Println("\nAutomation tip:")
	fmt.Println("If you have a test script, you can automate the process:")
	fmt.Println("git bisect run ./test.sh")
	fmt.Println("\nTo abort the bisect process:")
	fmt.Println("git bisect reset")

	return nil
}

func selectCommitForBisect() (string, error) {
	// Try using fzf if available
	if !noFzf {
		if _, err := exec.LookPath("fzf"); err == nil {
			return selectCommitWithFzfForBisect()
		}
	}
	return selectCommitWithListForBisect()
}

func selectCommitWithFzfForBisect() (string, error) {
	// Get git log
	logCmd := exec.Command("git", "log", "--oneline", "--color=always")
	logOutput, err := logCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git log: %w", err)
	}

	// Create fzf command with preview window showing commit details
	fzfCmd := exec.Command("fzf", 
		"--ansi",
		"--height", "50%",
		"--reverse",
		"--preview", "git show --color=always {1}",
		"--preview-window", "right:50%")
	fzfCmd.Stdin = strings.NewReader(string(logOutput))
	fzfCmd.Stderr = os.Stderr

	// Get fzf output
	output, err := fzfCmd.Output()
	if err != nil {
		return "", nil // User cancelled
	}

	// Extract commit hash
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", nil
	}

	return strings.Fields(selected)[0], nil
}

func selectCommitWithListForBisect() (string, error) {
	// Get recent commits
	logCmd := exec.Command("git", "log", "--oneline", "-n", "20")
	output, err := logCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git log: %w", err)
	}

	// Display commits
	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	fmt.Println("\nRecent commits:")
	for i, commit := range commits {
		fmt.Printf("%2d: %s\n", i+1, commit)
	}

	// Get user selection
	fmt.Print("\nSelect commit number (or press Enter to cancel): ")
	var input string
	fmt.Scanln(&input)

	if input == "" {
		return "", nil
	}

	var index int
	if _, err := fmt.Sscanf(input, "%d", &index); err != nil || index < 1 || index > len(commits) {
		return "", fmt.Errorf("invalid selection")
	}

	return strings.Fields(commits[index-1])[0], nil
} 