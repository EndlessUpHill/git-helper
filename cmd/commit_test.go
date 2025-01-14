package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupGitRepo(t *testing.T) (string, func()) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	assert.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	assert.NoError(t, cmd.Run())

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	assert.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	// Stage the file
	cmd = exec.Command("git", "add", "test.txt")
	cmd.Dir = tmpDir
	assert.NoError(t, cmd.Run())

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestCheckGitRepo(t *testing.T) {
	// Test with valid git repo
	tmpDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Change to the test directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	err := checkGitRepo()
	assert.NoError(t, err)

	// Test with non-git directory
	nonGitDir, err := os.MkdirTemp("", "non-git-*")
	assert.NoError(t, err)
	defer os.RemoveAll(nonGitDir)

	os.Chdir(nonGitDir)
	err = checkGitRepo()
	assert.Error(t, err)
}

func TestGetStagedChangesSummary(t *testing.T) {
	tmpDir, cleanup := setupGitRepo(t)
	defer cleanup()

	// Change to the test directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	summary, err := getStagedChangesSummary()
	assert.NoError(t, err)
	assert.Contains(t, summary, "test.txt")
}

func TestGenerateCommitMessage(t *testing.T) {
	tests := []struct {
		name       string
		summary    string
		commitType string
		useAI      bool
		wantErr    bool
	}{
		{
			name:       "manual commit without type",
			summary:    "test.txt | 1 +",
			commitType: "",
			useAI:      false,
			wantErr:    false,
		},
		{
			name:       "manual commit with type",
			summary:    "test.txt | 1 +",
			commitType: "feat",
			useAI:      false,
			wantErr:    false,
		},
		{
			name:       "AI commit without API key",
			summary:    "test.txt | 1 +",
			useAI:      true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test state
			useAI = tt.useAI
			commitType = tt.commitType

			msg, err := generateCommitMessage(tt.summary)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if !tt.useAI {
					assert.Contains(t, msg, tt.summary)
				}
			}
		})
	}
} 