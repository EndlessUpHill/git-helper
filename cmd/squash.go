package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/EndlessUphill/git-helper/internal/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	message string
)

var squashCmd = &cobra.Command{
	Use:   "squash [number]",
	Short: "Squash recent commits into one",
	Long: `Quickly squash your recent commits into a single commit.

This command helps clean up your commit history by:
1. Showing commits that will be squashed
2. Creating a new commit message (manual or AI-generated)
3. Safely combining the commits

Useful when:
- Your commit history is too granular
- You want to clean up WIP commits
- You need a clean history before merging

Example:
  githelper squash 3                    # Squash last 3 commits
  githelper squash 5 -m "New feature"   # Squash with custom message
  githelper squash 3 --ai               # Generate message with AI`,
	Args: cobra.ExactArgs(1),
	RunE: runSquash,
}

func init() {
	rootCmd.AddCommand(squashCmd)
	squashCmd.Flags().StringVarP(&message, "message", "m", "", "custom commit message for squashed commit")
	squashCmd.Flags().BoolVar(&useAI, "ai", false, "use AI to generate commit message")
}

func runSquash(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Parse number of commits
	numCommits, err := strconv.Atoi(args[0])
	if err != nil || numCommits < 2 {
		return fmt.Errorf("please provide a valid number of commits (minimum 2)")
	}

	// Show commits that will be squashed
	fmt.Printf("🔍 Last %d commits to be squashed:\n\n", numCommits)
	logCmd := exec.Command("git", "log", "-n", strconv.Itoa(numCommits), "--oneline")
	logCmd.Stdout = os.Stdout
	logCmd.Stderr = os.Stderr
	if err := logCmd.Run(); err != nil {
		return fmt.Errorf("failed to show commits: %w", err)
	}

	// Confirm action
	fmt.Printf("\n⚠️  This will squash the above %d commits into one!\n", numCommits)
	if !confirmAction() {
		fmt.Println("❌ Operation cancelled")
		return nil
	}

	// Get commit messages for AI or default message
	var commitMessages string
	if useAI || message == "" {
		msgs, err := getCommitMessages(numCommits)
		if err != nil {
			return err
		}
		commitMessages = msgs
	}

	// Prepare commit message
	var finalMessage string
	if message != "" {
		finalMessage = message
	} else if useAI {
		// Generate message using AI
		msg, err := generateSquashMessage(commitMessages)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}
		finalMessage = msg
	} else {
		// Create default message from commit messages
		finalMessage = fmt.Sprintf("squash: %s", createDefaultMessage(commitMessages))
	}

	// Perform soft reset
	fmt.Printf("\n🔄 Resetting last %d commits...\n", numCommits)
	resetCmd := exec.Command("git", "reset", "--soft", fmt.Sprintf("HEAD~%d", numCommits))
	resetCmd.Stderr = os.Stderr
	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("failed to reset commits: %w", err)
	}

	// Create new commit
	fmt.Println("📝 Creating new squashed commit...")
	commitCmd := exec.Command("git", "commit", "-m", finalMessage)
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to create squashed commit: %w", err)
	}

	fmt.Printf("✅ Successfully squashed %d commits!\n", numCommits)
	return nil
}

func getCommitMessages(num int) (string, error) {
	cmd := exec.Command("git", "log", "-n", strconv.Itoa(num), "--format=%B")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit messages: %w", err)
	}
	return string(output), nil
}

func createDefaultMessage(messages string) string {
	// Split messages into lines
	lines := strings.Split(strings.TrimSpace(messages), "\n")
	
	// Get first line of each commit
	var firstLines []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			firstLines = append(firstLines, line)
			if len(firstLines) >= 3 {
				break
			}
		}
	}

	// Join first three (or fewer) commit messages
	summary := strings.Join(firstLines, "; ")
	if len(firstLines) < len(lines) {
		summary += "..."
	}

	return summary
}

func generateSquashMessage(messages string) (string, error) {
	// If AI flag is enabled but OpenAI key is not configured
	if !viper.IsSet("openai_api_key") {
		return createDefaultMessage(messages), nil
	}

	// Get OpenAI API key
	apiKey := viper.GetString("openai_api_key")
	generator := ai.NewCommitGenerator(apiKey)

	// Generate commit message
	message, err := generator.GenerateCommitMessage(messages)
	if err != nil {
		return createDefaultMessage(messages), nil
	}

	return message, nil
} 