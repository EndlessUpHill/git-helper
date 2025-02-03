package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	numFiles  int
	threshold string
)

var cleanCmd = &cobra.Command{
	Use:   "clean [file]",
	Short: "Find and remove large files from git history",
	Long: `Find and remove large files that are bloating your git repository.

This command helps you clean up your repository by:
1. Finding the largest files in git history
2. Letting you select which files to remove
3. Completely removing selected files from git history
4. Optionally force pushing the cleaned history

⚠️  WARNING: This rewrites git history! Use with caution on shared repositories.

Example:
  githelper clean              # Interactive file selection
  githelper clean large.zip   # Remove specific file
  githelper clean --top 20    # Show top 20 largest files
  githelper clean --min 100MB # Show files larger than 100MB`,
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().IntVarP(&numFiles, "top", "n", 10, "number of largest files to show")
	cleanCmd.Flags().StringVarP(&threshold, "min", "m", "", "minimum file size (e.g., 100MB)")
}

type LargeFile struct {
	Path string
	Size int64
}

func runClean(cmd *cobra.Command, args []string) error {
	if err := checkGitRepo(); err != nil {
		return err
	}

	var fileToPurge string
	var err error

	if len(args) > 0 {
		fileToPurge = args[0]
	} else {
		// Find and select large file
		fmt.Println("🔍 Finding large files in git history...")
		fileToPurge, err = selectLargeFile()
		if err != nil {
			return err
		}
		if fileToPurge == "" {
			return fmt.Errorf("no file selected")
		}
	}

	// Confirm action
	fmt.Printf("\n⚠️  WARNING: This will permanently remove '%s' from git history!\n", fileToPurge)
	fmt.Println("This action CANNOT be undone and will rewrite git history.")
	if !confirmAction() {
		fmt.Println("❌ Operation cancelled")
		return nil
	}

	// Remove file from git history
	fmt.Printf("\n🗑️  Removing '%s' from history...\n", fileToPurge)
	filterCmd := exec.Command("git", "filter-branch", "--force",
		"--index-filter", fmt.Sprintf("git rm --cached --ignore-unmatch %s", fileToPurge),
		"--prune-empty", "--tag-name-filter", "cat", "--", "--all")
	
	filterCmd.Stdout = os.Stdout
	filterCmd.Stderr = os.Stderr
	
	if err := filterCmd.Run(); err != nil {
		return fmt.Errorf("failed to remove file from history: %w", err)
	}

	fmt.Println("\n✅ File removed from git history!")
	fmt.Println("\n⚠️  To push these changes:")
	fmt.Println("git push origin --force --all")

	return nil
}

func selectLargeFile() (string, error) {
	// Try using fzf if available
	if !noFzf {
		if _, err := exec.LookPath("fzf"); err == nil {
			return selectLargeFileWithFzf()
		}
	}
	return selectLargeFileWithList()
}

func getLargeFiles() ([]LargeFile, error) {
	// Get all objects in git history
	cmd := exec.Command("sh", "-c", `git rev-list --objects --all | awk '{print $1}' | git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' | grep '^blob' | awk '{print $3 " " $4}'`)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git objects: %w", err)
	}

	// Parse output and create file list
	var files []LargeFile
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 2)
		if len(parts) != 2 {
			continue
		}

		size, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		// Apply size threshold if specified
		if threshold != "" {
			thresholdBytes, err := parseSize(threshold)
			if err != nil {
				return nil, fmt.Errorf("invalid size threshold: %w", err)
			}
			if size < thresholdBytes {
				continue
			}
		}

		files = append(files, LargeFile{
			Path: parts[1],
			Size: size,
		})
	}

	// Sort by size
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	// Limit to top N files
	if len(files) > numFiles {
		files = files[:numFiles]
	}

	return files, nil
}

func selectLargeFileWithFzf() (string, error) {
	files, err := getLargeFiles()
	if err != nil {
		return "", err
	}

	// Create input for fzf
	var input strings.Builder
	for _, file := range files {
		fmt.Fprintf(&input, "%s (%s)\n", file.Path, formatSize(file.Size))
	}

	// Create preview command
	previewCmd := "git log --oneline --all -- {1}"

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

	// Extract filename from selection
	selected := strings.TrimSpace(string(output))
	return strings.Fields(selected)[0], nil
}

func selectLargeFileWithList() (string, error) {
	files, err := getLargeFiles()
	if err != nil {
		return "", err
	}

	fmt.Println("\nLargest files in repository:")
	for i, file := range files {
		fmt.Printf("%2d: %s (%s)\n", i+1, file.Path, formatSize(file.Size))
	}

	fmt.Print("\nSelect file number (or press Enter to cancel): ")
	var input string
	fmt.Scanln(&input)

	if input == "" {
		return "", nil
	}

	var index int
	if _, err := fmt.Sscanf(input, "%d", &index); err != nil || index < 1 || index > len(files) {
		return "", fmt.Errorf("invalid selection")
	}

	return files[index-1].Path, nil
}

func parseSize(size string) (int64, error) {
	size = strings.ToUpper(size)
	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	for suffix, multiplier := range multipliers {
		if strings.HasSuffix(size, suffix) {
			value := strings.TrimSuffix(size, suffix)
			number, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return 0, err
			}
			return int64(number * float64(multiplier)), nil
		}
	}

	// Try parsing as bytes if no suffix
	bytes, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %s", size)
	}
	return bytes, nil
}

