package git

import (
	"os/exec"
	"path/filepath"
)

// GetGitRoot returns the absolute path to the git repository root
func GetGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return filepath.Clean(string(output)), nil
} 