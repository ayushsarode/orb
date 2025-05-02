// internal/cmd/commit.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ayushsarode/orb/internal/config"
	"github.com/ayushsarode/orb/internal/index"
	"github.com/ayushsarode/orb/internal/objects"
	"github.com/ayushsarode/orb/internal/refs"
	"github.com/spf13/cobra"
)

func newCommitCommand() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Record changes to the repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if message == "" {
				return fmt.Errorf("commit message cannot be empty")
			}

			idx, err := index.LoadIndex()
			if err != nil {
				return fmt.Errorf("loading index: %w", err)
			}

			if len(idx.Entries) == 0 {
				return fmt.Errorf("nothing to commit")
			}

			// Create a tree object from the index
			treeContent := buildTreeContent(idx)
			treeHash, err := objects.WriteObject(objects.TreeType, treeContent)
			if err != nil {
				return fmt.Errorf("writing tree object: %w", err)
			}

			// get the current HEAD commit (if any)
			// if im on a branch, the current branch commit is my parent commit here
			parentHash := ""
			head, err := refs.GetHead()
			if err == nil {
				parentHash = head
			}

			// Build commit content
			commitContent := buildCommitContent(treeHash, parentHash, message)
			commitHash, err := objects.WriteObject(objects.CommitType, commitContent)
			if err != nil {
				return fmt.Errorf("writing commit object: %w", err)
			}

			// Update the current branch (or HEAD if detached)
			headContent, err := os.ReadFile(".orb/HEAD")
			if err != nil {
				return fmt.Errorf("reading HEAD: %w", err)
			}

			headRef := string(headContent)
			if len(headRef) > 5 && headRef[:5] == "ref: " {
				// HEAD points to a branch
				branchRef := headRef[5 : len(headRef)-1] // Remove "ref: " and newline
				if err := refs.UpdateRef(branchRef, commitHash); err != nil {
					return fmt.Errorf("updating branch reference: %w", err)
				}
			} else {
				// Detached HEAD
				if err := refs.UpdateHead(commitHash); err != nil {
					return fmt.Errorf("updating HEAD: %w", err)
				}
			}

			fmt.Printf("[%s] %s\n", commitHash[:7], message)
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message")
	cmd.MarkFlagRequired("message")

	return cmd
}

// buildTreeContent creates the content for a tree object from the index
func buildTreeContent(idx *index.Index) []byte {

	dirMap := make(map[string][]index.Entry)

	for _, entry := range idx.GetEntries() {
		dir := filepath.Dir(entry.Path)
		dirMap[dir] = append(dirMap[dir], entry)
	}

	rootTree := buildDirectoryTree(dirMap, ".")
	return rootTree

}

func buildDirectoryTree(dirMap map[string][]index.Entry, path string) []byte {
	var result []byte

	// Add entries for this directory
	// Loop over every entry (file) in this folder (path)
	for _, entry := range dirMap[path] {
		if filepath.Dir(entry.Path) != path {
			continue
		}

		// Format: "mode type hash\tname" (basename only, not full path)
		name := filepath.Base(entry.Path)
		line := fmt.Sprintf("100644 blob %s\t%s", entry.ObjectHash, name)
		result = append(result, []byte(line)...)
		result = append(result, '\n')
	}

	// Add subdirectories as tree objects
	for dir := range dirMap {
		if dir != path && filepath.Dir(dir) == path {
			// Create a tree for this subdirectory
			subTree := buildDirectoryTree(dirMap, dir)
			subTreeHash, _ := objects.WriteObject(objects.TreeType, subTree)

			// Add entry for this subdirectory
			name := filepath.Base(dir)
			line := fmt.Sprintf("040000 tree %s\t%s", subTreeHash, name)
			result = append(result, []byte(line)...)
			result = append(result, '\n')
		}
	}

	return result
}

func buildCommitContent(treeHash, parentHash, message string) []byte {
	var content string

	content += fmt.Sprintf("tree %s\n", treeHash)
	if parentHash != "" {
		content += fmt.Sprintf("parent %s\n", parentHash)
	}

	// Load user configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		// Fallback to defaults if config can't be loaded
		cfg = &config.Config{Values: make(map[string]string)}
	}

	// Get user name and email from config, with defaults
	authorName := cfg.GetString("user.name", "Unknown")
	authorEmail := cfg.GetString("user.email", "unknown@example.com")

	// Get local timezone
	now := time.Now()
	_, tzOffsetSeconds := now.Zone()

	// Format timezone as ±HHMM
	tzHours := tzOffsetSeconds / 3600
	tzMinutes := (tzOffsetSeconds % 3600) / 60
	timezone := fmt.Sprintf("%+03d%02d", tzHours, tzMinutes)

	timestamp := now.Unix()

	author := fmt.Sprintf("%s <%s> %d %s", authorName, authorEmail, timestamp, timezone)
	committer := author // Use same info for committer

	content += fmt.Sprintf("author %s\n", author)
	content += fmt.Sprintf("committer %s\n", committer)
	content += fmt.Sprintf("\n%s\n", message)

	return []byte(content)
}
