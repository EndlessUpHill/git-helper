package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/EndlessUphill/git-helper/internal/ai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	skipEdit    bool
	commitType  string
	useAI      bool
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate and make a conventional commit",
	Long: `Generate a commit message following conventional commit standards.
Common types: feat, fix, docs, style, refactor, test, chore.
Format: <type>[optional scope]: <description>

Example: feat(auth): add OAuth2 authentication`,
	RunE: runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)
	flags := commitCmd.Flags()
	flags.BoolVarP(&skipEdit, "no-edit", "n", false, "skip editing the generated message")
	flags.StringVarP(&commitType, "type", "t", "", "commit type (feat, fix, docs, etc.)")
	flags.BoolVarP(&useAI, "ai", "a", false, "use AI to generate commit message")
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Check if current directory is a git repository
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Get staged changes summary
	summary, err := getStagedChangesSummary()
	if err != nil {
		return err
	}

	if summary == "" {
		return fmt.Errorf("no staged changes found. Use 'git add' to stage changes")
	}

	// Generate commit message
	message, err := generateCommitMessage(summary)
	if err != nil {
		return err
	}

	// Allow user to edit unless --no-edit flag is set
	if !skipEdit {
		message, err = editMessage(message)
		if err != nil {
			return err
		}
	}

	// Make the commit
	return makeCommit(message)
}

func checkGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a git repository")
	}
	return nil
}

func getStagedChangesSummary() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--stat")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get staged changes: %w", err)
	}
	return string(output), nil
}

func getDetailedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get detailed diff: %w", err)
	}
	return string(output), nil
}

func generateCommitMessage(summary string) (string, error) {
	var message strings.Builder

	if useAI {
		// Get detailed diff for AI
		diff, err := getDetailedDiff()
		if err != nil {
			return "", err
		}

		// Get OpenAI API key
		apiKey := viper.GetString("openai_api_key")
		if apiKey == "" {
			return "", fmt.Errorf("OpenAI API key not found in config")
		}

		// Generate commit message using AI
		generator := ai.NewCommitGenerator(apiKey)
		aiMessage, err := generator.GenerateCommitMessage(diff)
		if err != nil {
			return "", err
		}

		message.WriteString(aiMessage)
	} else {
		// Original manual commit message generation
		if commitType == "" {
			fmt.Println("Available commit types:")
			fmt.Println("1. feat     - A new feature")
			fmt.Println("2. fix      - A bug fix")
			fmt.Println("3. docs     - Documentation only changes")
			fmt.Println("4. style    - Changes that don't affect the meaning of the code")
			fmt.Println("5. refactor - Code change that neither fixes a bug nor adds a feature")
			fmt.Println("6. test     - Adding missing tests or correcting existing tests")
			fmt.Println("7. chore    - Changes to the build process or auxiliary tools")
			
			fmt.Print("\nEnter commit type (or number): ")
			var input string
			fmt.Scanln(&input)

			// Handle numeric input
			switch input {
			case "1":
				commitType = "feat"
			case "2":
				commitType = "fix"
			case "3":
				commitType = "docs"
			case "4":
				commitType = "style"
			case "5":
				commitType = "refactor"
			case "6":
				commitType = "test"
			case "7":
				commitType = "chore"
			default:
				commitType = input
			}
		}
		message.WriteString(fmt.Sprintf("%s: ", commitType))
	}

	// Add summary of changes
	message.WriteString("\n\n# Changes to be committed:\n")
	message.WriteString(fmt.Sprintf("# %s\n", summary))
	if useAI {
		message.WriteString("\n# AI-generated commit message above\n")
	}
	message.WriteString("# Lines starting with '#' will be ignored\n")

	return message.String(), nil
}

func editMessage(message string) (string, error) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "COMMIT_EDITMSG")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write message to file
	if _, err := tmpfile.WriteString(message); err != nil {
		return "", fmt.Errorf("failed to write to temporary file: %w", err)
	}
	tmpfile.Close()

	// Get editor command (use EDITOR env var or default to vim)
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Open editor
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to open editor: %w", err)
	}

	// Read edited message
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited message: %w", err)
	}

	// Remove comments and empty lines
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(strings.TrimSpace(line), "#") && line != "" {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n"), nil
}

func makeCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
} 