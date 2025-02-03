package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	
	forceMode bool
)

var pruneRemotesCmd = &cobra.Command{
	Use:   "prune-remotes",
	Short: "Detect and remove unused Git remotes",
	Long: `Clean up your Git remotes by removing unreachable ones.

This command helps manage your Git remotes by:
1. Listing all configured remotes
2. Testing connectivity to each remote
3. Removing unreachable remotes

Useful when:
- You have old remotes that are no longer needed
- Remote repositories have been deleted or moved
- You want to clean up your remote configuration

Example:
  githelper prune-remotes         # Interactive remote cleanup
  githelper prune-remotes --dry   # Show what would be removed
  githelper prune-remotes --force # Remove without confirmation`,
	RunE: runPruneRemotes,
}

func init() {
	rootCmd.AddCommand(pruneRemotesCmd)
	pruneRemotesCmd.Flags().BoolVar(&dryRun, "dry", false, "dry run (show what would be removed)")
	pruneRemotesCmd.Flags().BoolVar(&forceMode, "force", false, "remove without confirmation")
}

type Remote struct {
	Name     string
	URL      string
	Reachable bool
}

func runPruneRemotes(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Get list of remotes
	remotes, err := getRemotes()
	if err != nil {
		return err
	}

	if len(remotes) == 0 {
		fmt.Println("No Git remotes configured.")
		return nil
	}

	// Check each remote
	fmt.Println("🔍 Checking remotes...")
	for i := range remotes {
		remotes[i].Reachable = checkRemote(remotes[i].Name)
	}

	// Show status
	unreachable := listUnreachableRemotes(remotes)
	if len(unreachable) == 0 {
		fmt.Println("✅ All remotes are reachable!")
		return nil
	}

	if dryRun {
		fmt.Println("\nThe following remotes would be removed:")
		for _, remote := range unreachable {
			fmt.Printf("- %s (%s)\n", remote.Name, remote.URL)
		}
		return nil
	}

	// Confirm removal
	if !forceMode {
		fmt.Println("\n⚠️  The following remotes will be removed:")
		for _, remote := range unreachable {
			fmt.Printf("- %s (%s)\n", remote.Name, remote.URL)
		}
		if !confirmAction() {
			fmt.Println("❌ Operation cancelled")
			return nil
		}
	}

	// Remove unreachable remotes
	removed := 0
	for _, remote := range unreachable {
		if err := removeRemote(remote.Name); err != nil {
			fmt.Printf("⚠️  Failed to remove remote '%s': %v\n", remote.Name, err)
			continue
		}
		removed++
		fmt.Printf("🗑️  Removed remote '%s'\n", remote.Name)
	}

	fmt.Printf("\n✅ Removed %d unreachable remote(s)\n", removed)
	return nil
}

func getRemotes() ([]Remote, error) {
	cmd := exec.Command("git", "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	remotes := make(map[string]string)

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.HasSuffix(line, "(fetch)") {
			remotes[parts[0]] = parts[1]
		}
	}

	var result []Remote
	for name, url := range remotes {
		result = append(result, Remote{
			Name: name,
			URL:  url,
		})
	}

	return result, nil
}

func checkRemote(name string) bool {
	cmd := exec.Command("git", "ls-remote", "--exit-code", name)
	cmd.Stderr = os.Stderr
	return cmd.Run() == nil
}

func listUnreachableRemotes(remotes []Remote) []Remote {
	var unreachable []Remote
	for _, remote := range remotes {
		if !remote.Reachable {
			unreachable = append(unreachable, remote)
		}
	}
	return unreachable
}

func removeRemote(name string) error {
	cmd := exec.Command("git", "remote", "remove", name)
	cmd.Stderr = os.Stderr
	return cmd.Run()
} 