package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/EndlessUphill/git-helper/internal/github"
	gh "github.com/google/go-github/v53/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	destination string
	isOrg       bool
	repoConfig  github.RepoConfig
)

var copyCmd = &cobra.Command{
	Use:   "copy [source-repo-url]",
	Short: "Copy a repository with full history",
	Long: `Copy a repository including all branches and tags to a new destination.
Example: githelper copy https://github.com/user/repo --dest newuser/repo`,
	Args: cobra.ExactArgs(1),
	RunE: runCopy,
}

func init() {
	rootCmd.AddCommand(copyCmd)
	flags := copyCmd.Flags()
	flags.StringVarP(&destination, "dest", "d", "", "destination repo (format: user/repo or org/repo)")
	flags.BoolVarP(&isOrg, "org", "o", false, "destination is an organization")
	flags.BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
	
	// Repository settings
	flags.BoolVar(&repoConfig.Private, "private", true, "make repository private")
	flags.StringVar(&repoConfig.Description, "description", "", "repository description")
	flags.StringSliceVar(&repoConfig.Topics, "topics", nil, "repository topics")
	flags.BoolVar(&repoConfig.HasIssues, "issues", true, "enable issues")
	flags.BoolVar(&repoConfig.HasWiki, "wiki", true, "enable wiki")
	
	// Add SSH option
	flags.Bool("ssh", true, "use SSH for git operations (default is HTTPS)")
	viper.BindPFlag("use_ssh", flags.Lookup("ssh"))
	
	copyCmd.MarkFlagRequired("dest")
}

func runCopy(cmd *cobra.Command, args []string) error {
	sourceURL := args[0]
	
	// Validate GitHub URL format
	_, err := parseGitHubURL(sourceURL)
	if err != nil {
		return err
	}

	if dryRun {
		return performDryRun(sourceURL, destination)
	}

	fmt.Printf("🔄 Starting repository copy from %s to %s\n", sourceURL, destination)

	// Get system temp directory
	tmpDir := os.TempDir()
	if tmpDir == "" {
		return fmt.Errorf("unable to determine system temporary directory")
	}

	// Create a subdirectory for our operation
	workDir, err := os.MkdirTemp(tmpDir, "githelper-copy-*")
	if err != nil {
		return fmt.Errorf("failed to create working directory in %s: %w", tmpDir, err)
	}
	defer func() {
		if err := os.RemoveAll(workDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean up temporary directory %s: %v\n", workDir, err)
		}
	}()

	fmt.Printf("📁 Working directory: %s\n", workDir)

	// Clone the source repository with mirror flag
	fmt.Printf("📥 Cloning source repository...\n")
	if err := cloneMirror(sourceURL, workDir); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("git clone failed: %s", exitErr.Stderr)
		}
		return fmt.Errorf("failed to clone source repository: %w", err)
	}

	// Create the destination repository
	fmt.Printf("📝 Creating destination repository...\n")
	if err := createDestinationRepo(destination, isOrg); err != nil {
		if ghErr, ok := err.(*gh.ErrorResponse); ok {
			if ghErr.Response.StatusCode == 422 {
				return fmt.Errorf("repository already exists or name is invalid")
			}
		}
		return fmt.Errorf("failed to create destination repository: %w", err)
	}

	// Push to destination
	fmt.Printf("📤 Pushing repository content...\n")
	if err := pushMirror(workDir, destination); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("git push failed: %s", exitErr.Stderr)
		}
		return fmt.Errorf("failed to push to destination: %w", err)
	}

	fmt.Printf("✅ Successfully copied repository to %s\n", destination)
	return nil
}

func performDryRun(sourceURL, dest string) error {
	fmt.Println("🔍 Dry run - no changes will be made")
	fmt.Printf("Would perform the following actions:\n\n")
	fmt.Printf("1. Create temporary directory for cloning\n")
	fmt.Printf("2. Clone %s with --mirror flag\n", sourceURL)
	fmt.Printf("3. Create new repository at %s\n", dest)
	fmt.Printf("   - Private: %v\n", repoConfig.Private)
	fmt.Printf("   - Description: %s\n", repoConfig.Description)
	if len(repoConfig.Topics) > 0 {
		fmt.Printf("   - Topics: %s\n", strings.Join(repoConfig.Topics, ", "))
	}
	fmt.Printf("   - Issues enabled: %v\n", repoConfig.HasIssues)
	fmt.Printf("   - Wiki enabled: %v\n", repoConfig.HasWiki)
	fmt.Printf("4. Push mirror to destination\n")
	fmt.Printf("5. Clean up temporary directory\n")
	return nil
}

func cloneMirror(sourceURL, dir string) error {
	cmd := exec.Command("git", "clone", "--mirror", sourceURL, dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createDestinationRepo(dest string, isOrg bool) error {
	ctx := context.Background()
	
	// Get GitHub token with more verbose error handling
	token := viper.GetString("github_token")
	if token == "" {
		// Try environment variable directly as fallback
		token = os.Getenv("GITHELPER_GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("GitHub token not found. Either:\n" +
				"1. Set GITHELPER_GITHUB_TOKEN environment variable\n" +
				"2. Add github_token to ~/.githelper.yaml\n" +
				"3. Use --config to specify a config file")
		}
	}

	if viper.GetBool("debug") {
		fmt.Printf("Token length: %d\n", len(token))
	}

	// Create our internal GitHub client
	client := github.NewClient(token)
	
	// Parse owner and repo name from destination
	owner, repo, found := strings.Cut(dest, "/")
	if !found {
		return fmt.Errorf("invalid destination format. Use 'owner/repo'")
	}

	// If description is empty, set a default one
	if repoConfig.Description == "" {
		repoConfig.Description = "Repository copied using GitHelper"
	}

	return client.CreateRepository(ctx, repo, owner, isOrg, repoConfig)
}

func pushMirror(dir, dest string) error {
	// Allow users to choose their preferred URL format
	useSSH := viper.GetBool("use_ssh")
	var destURL string
	
	if useSSH {
		destURL = fmt.Sprintf("git@github.com:%s.git", dest)
	} else {
		destURL = fmt.Sprintf("https://github.com/%s.git", dest)
	}

	cmd := exec.Command("git", "push", "--mirror", destURL)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Add this function to parse and validate GitHub URLs
func parseGitHubURL(url string) (string, error) {
	// Handle SSH format: git@github.com:user/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.TrimPrefix(url, "git@github.com:")
		url = strings.TrimSuffix(url, ".git")
		return url, nil
	}

	// Handle HTTPS format: https://github.com/user/repo
	if strings.HasPrefix(url, "https://github.com/") {
		url = strings.TrimPrefix(url, "https://github.com/")
		url = strings.TrimSuffix(url, ".git")
		return url, nil
	}

	return "", fmt.Errorf("invalid GitHub URL format. Use HTTPS (https://github.com/user/repo) or SSH (git@github.com:user/repo)")
} 