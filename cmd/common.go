package cmd

import "fmt"

type ReflogEntry struct {
	Hash        string
	Action      string
	Description string
}
	
var (
	mainBranch string
	force      bool
	dryRun     bool
	useAI      bool
)

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
} 