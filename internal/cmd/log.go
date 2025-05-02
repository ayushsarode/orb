// internal/cmd/log.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ayushsarode/orb/internal/objects"
	"github.com/ayushsarode/orb/internal/refs"
	"github.com/spf13/cobra"
)

func newLogCommand() *cobra.Command {
	var quiet bool

	cmd := &cobra.Command{
		Use:   "log",
		Short: "Show commit logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			startRef := "HEAD"
			if len(args) > 0 {
				startRef = args[0]
			}

			// Get the commit hash
			commitHash, err := refs.GetRef(startRef)
			if err != nil {
				return fmt.Errorf("getting reference: %w", err)
			}

			// Create a map of commit hashes to branch names
			branchRefs := make(map[string][]string)

			// Read all branches
			branchFiles, err := os.ReadDir(refs.HeadsDir)
			if err == nil {
				for _, file := range branchFiles {
					if file.IsDir() {
						continue
					}

					branchName := file.Name()
					branchPath := filepath.Join(refs.HeadsDir, branchName)
					branchCommit, err := os.ReadFile(branchPath)
					if err != nil {
						continue
					}

					hash := strings.TrimSpace(string(branchCommit))
					branchRefs[hash] = append(branchRefs[hash], branchName)
				}
			}

			// commit history
			for commitHash != "" {
				objType, content, err := objects.ReadObject(commitHash)
				if err != nil {
					fmt.Printf("Warning: Error reading commit %s: %v\n", commitHash, err)
					if quiet {
						return nil
					}
					break
				}

				if objType != objects.CommitType {
					fmt.Printf("Warning: Object is not a commit: %s\n", commitHash)
					break
				}

				// parse the commit
				commit, err := parseCommit(string(content))
				if err != nil {
					fmt.Printf("Warning: Error parsing commit: %v\n", err)
					break
				}

				fmt.Printf("commit %s", commitHash)

				// branch information if this commit is the head of any branches
				if branches, ok := branchRefs[commitHash]; ok && len(branches) > 0 {
					fmt.Printf(" (")
					for i, branch := range branches {
						if i > 0 {
							fmt.Printf(", ")
						}

						// Highlight current branch
						headContent, _ := os.ReadFile(refs.HeadFile)
						currentBranch := ""
						if headRef := strings.TrimSpace(string(headContent)); strings.HasPrefix(headRef, "ref: refs/heads/") {
							currentBranch = strings.TrimPrefix(headRef, "ref: refs/heads/")
						}

						if branch == currentBranch {
							fmt.Printf("HEAD -> %s", branch)
						} else {
							fmt.Printf("%s", branch)
						}
					}
					fmt.Printf(")")
				}
				fmt.Printf("\n")

				fmt.Printf("Author: %s\n", commit.Author)

				// Only format date if it's not zero
				if !commit.CommitTime.IsZero() {
					fmt.Printf("Date:   %s\n", commit.CommitTime.Format(time.RFC1123Z))
				} else {
					fmt.Printf("Date:   (unknown)\n")
				}

				fmt.Printf("\n    %s\n\n", commit.Message)

				// check if parent exists before trying to access it
				if commit.Parent == "" {
					break
				}

				// Check if parent object exists before continuing
				parentExists := true
				parentPath := filepath.Join(".orb/objects", commit.Parent[:2], commit.Parent[2:])
				if _, err := os.Stat(parentPath); os.IsNotExist(err) {
					if !quiet {
						fmt.Printf("Warning: Parent commit %s not found\n", commit.Parent)
					}
					parentExists = false
				}

				if !parentExists {
					break
				}

				commitHash = commit.Parent
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress warnings about missing commits")

	return cmd
}

// Commit represents a parsed commit object
type Commit struct {
	TreeHash   string
	Parent     string
	Author     string
	Committer  string
	Message    string
	CommitTime time.Time
}

// parseCommit parses the content of a commit object
func parseCommit(content string) (*Commit, error) {
	commit := &Commit{}

	lines := strings.Split(content, "\n")
	messageStart := -1

	for i, line := range lines {
		if line == "" {
			messageStart = i + 1
			break
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "tree":
			commit.TreeHash = value
		case "parent":
			commit.Parent = value
		case "author":
			// Extract name and email part, removing the timestamp
			if idx := strings.LastIndex(value, ">"); idx > 0 {
				commit.Author = value[:idx+1]

				// Parse timestamp for display in the Date field
				timestampStr := strings.TrimSpace(value[idx+1:])
				if ts, err := parseTimestamp(timestampStr); err == nil {
					commit.CommitTime = ts
				}
			} else {
				commit.Author = value
			}
		case "committer":
			commit.Committer = value
		}
	}

	// Extract the commit message
	if messageStart >= 0 && messageStart < len(lines) {
		commit.Message = strings.Join(lines[messageStart:], "\n")
	}

	return commit, nil
}

// parseTimestamp parses a Git-style timestamp (e.g. "1621234567 +0200") into a time.Time
func parseTimestamp(timestamp string) (time.Time, error) {
	parts := strings.Split(timestamp, " ")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid timestamp format")
	}

	unixTime, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid unix timestamp: %w", err)
	}

	return time.Unix(unixTime, 0), nil
}
