package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	showAll bool
	sortBy  string
)

var branchSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Interactively switch between Git branches",
	Long: `Interactive branch switching with search capabilities.

This command helps you quickly switch between branches:
1. Shows list of branches sorted by last commit
2. Provides interactive search with preview
3. Switches to selected branch instantly

Useful when:
- Working across multiple branches
- Need to find a specific branch quickly
- Want to see branch details before switching

Example:
  githelper switch           # Interactive branch selection
  githelper switch --all    # Show all branches (including remote)
  githelper switch --sort=name  # Sort by branch name`,
	RunE: runSwitch,
}

type Branch struct {
	Name           string
	LastCommitHash string
	LastCommitDate time.Time
	LastCommitMsg  string
	Current        bool
}

func init() {
	rootCmd.AddCommand(branchSwitchCmd)
	branchSwitchCmd.Flags().BoolVar(&showAll, "all", false, "show all branches (including remote)")
	branchSwitchCmd.Flags().StringVar(&sortBy, "sort", "date", "sort by: date, name")
}

func runSwitch(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	// Check for uncommitted changes
	if hasChanges, err := hasUncommittedChanges(); err != nil {
		return err
	} else if hasChanges {
		return fmt.Errorf("you have uncommitted changes. Please commit or stash them first")
	}

	// Get branches
	branches, err := getBranches()
	if err != nil {
		return err
	}

	if len(branches) == 0 {
		return fmt.Errorf("no branches found")
	}

	// Select branch
	selected, err := selectBranch(branches)
	if err != nil {
		return err
	}
	if selected == "" {
		return fmt.Errorf("no branch selected")
	}

	// Switch to branch
	fmt.Printf("🔄 Switching to branch '%s'...\n", selected)
	checkoutCmd := exec.Command("git", "checkout", selected)
	checkoutCmd.Stdout = os.Stdout
	checkoutCmd.Stderr = os.Stderr
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to switch branch: %w", err)
	}

	fmt.Printf("✅ Switched to branch '%s'\n", selected)
	return nil
}

func getBranches() ([]Branch, error) {
	var args []string
	if showAll {
		args = []string{"branch", "-a", "--format", "%(refname:short) %(objectname) %(committerdate:iso) %(contents:subject)"}
	} else {
		args = []string{"branch", "--format", "%(refname:short) %(objectname) %(committerdate:iso) %(contents:subject)"}
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var branches []Branch
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 4)
		if len(parts) < 4 {
			continue
		}

		name := parts[0]
		hash := parts[1]
		dateStr := parts[2]
		msg := parts[3]

		date, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
		if err != nil {
			date = time.Time{}
		}

		branches = append(branches, Branch{
			Name:           name,
			LastCommitHash: hash,
			LastCommitDate: date,
			LastCommitMsg:  msg,
			Current:        strings.HasPrefix(name, "* "),
		})
	}

	// Sort branches
	switch sortBy {
	case "name":
		sortBranchesByName(branches)
	default:
		sortBranchesByDate(branches)
	}

	return branches, nil
}

func selectBranch(branches []Branch) (string, error) {
	if !noFzf {
		if _, err := exec.LookPath("fzf"); err == nil {
			return selectBranchWithFzf(branches)
		}
	}
	return selectBranchWithList(branches)
}

func selectBranchWithFzf(branches []Branch) (string, error) {
	// Create input for fzf
	var input strings.Builder
	for _, branch := range branches {
		fmt.Fprintf(&input, "%s\t%s\t%s\n",
			branch.Name,
			branch.LastCommitDate.Format("2006-01-02 15:04:05"),
			branch.LastCommitMsg)
	}

	// Create preview command that shows branch details
	previewCmd := "git log --color=always --oneline --graph {1}"

	fzfCmd := exec.Command("fzf",
		"--ansi",
		"--height", "50%",
		"--reverse",
		"--preview", previewCmd,
		"--preview-window", "right:50%",
		"--with-nth", "1,2,3",
		"--delimiter", "\t")

	fzfCmd.Stdin = strings.NewReader(input.String())
	fzfCmd.Stderr = os.Stderr

	output, err := fzfCmd.Output()
	if err != nil {
		return "", nil // User cancelled
	}

	// Extract branch name from selection
	selected := strings.TrimSpace(string(output))
	return strings.Fields(selected)[0], nil
}

func selectBranchWithList(branches []Branch) (string, error) {
	fmt.Println("\nAvailable branches:")
	for i, branch := range branches {
		fmt.Printf("%2d: %s (%s) - %s\n",
			i+1,
			branch.Name,
			branch.LastCommitDate.Format("2006-01-02"),
			branch.LastCommitMsg)
	}

	fmt.Print("\nSelect branch number (or press Enter to cancel): ")
	var input string
	fmt.Scanln(&input)

	if input == "" {
		return "", nil
	}

	var index int
	if _, err := fmt.Sscanf(input, "%d", &index); err != nil || index < 1 || index > len(branches) {
		return "", fmt.Errorf("invalid selection")
	}

	return branches[index-1].Name, nil
}

func sortBranchesByDate(branches []Branch) {
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].LastCommitDate.After(branches[j].LastCommitDate)
	})
}

func sortBranchesByName(branches []Branch) {
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].Name < branches[j].Name
	})
} 