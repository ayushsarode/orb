package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ayushsarode/orb/internal/index"
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the working tree status",
		RunE: func(cmd *cobra.Command, args []string) error {

			idx, err := index.LoadIndex()
			if err != nil {
				return fmt.Errorf("loading index: %w", err)
			}																																																		

			// Get all files in the working directory
			workingFiles, err := listWorkingDirFiles()
			if err != nil {
				return fmt.Errorf("listing files: %w", err)
			}

			// tracking
			stagedFiles := make(map[string]bool)
			modifiedFiles := make(map[string]bool)
			untrackedFiles := make(map[string]bool)

			// check which files are in the index
			for path := range idx.Entries {
				stagedFiles[path] = true

				// check if the file still exists
				if _, err := os.Stat(path); os.IsNotExist(err) {
					modifiedFiles[path] = true
					continue
				}

				// check if file is modified
				modified, err := isFileModified(path, idx.Entries[path])
				if err != nil {
					return fmt.Errorf("checking if file is modified: %w", err)
				}

				if modified {
					modifiedFiles[path] = true
				}
			}

			// Find untracked files
			for _, path := range workingFiles {
				if !stagedFiles[path] {
					untrackedFiles[path] = true
				}
			}

			// Get current 1 name
			headContent, err := os.ReadFile(".orb/HEAD")
			branchName := "detached HEAD"

			if err == nil {
				head := strings.TrimSpace(string(headContent))
				if strings.HasPrefix(head, "ref: refs/heads/") {
					// Use the refs package to get the branch name
					refPath := strings.TrimPrefix(head, "ref: ")
					if strings.HasPrefix(refPath, "refs/heads/") {
						branchName = strings.TrimPrefix(refPath, "refs/heads/")
					}
				} else {
					// We're in detached HEAD state with a commit hash
					shortHash := head
					if len(head) > 7 {
						shortHash = head[:7]
					}
					branchName = fmt.Sprintf("detached HEAD at %s", shortHash)
				}
			}

			// Print status with correct branch name
			fmt.Printf("On branch %s\n", branchName)

			// Print staged files
			if len(stagedFiles) > 0 {
				fmt.Println("\nChanges to be committed:")
				fmt.Println("  (use \"orb reset HEAD <file>...\" to unstage)")

				for path := range stagedFiles {
					if _, err := os.Stat(path); os.IsNotExist(err) {
						fmt.Printf("\tdeleted:    %s\n", path)
					} else {
						fmt.Printf("\tnew file:   %s\n", path)
					}
				}
			}

			// Print modified files
			if len(modifiedFiles) > 0 {
				fmt.Println("\nChanges not staged for commit:")
				fmt.Println("  (use \"orb add <file>...\" to update what will be committed)")
				fmt.Println("  (use \"orb checkout -- <file>...\" to discard changes in working directory)")

				for path := range modifiedFiles {
					if _, err := os.Stat(path); os.IsNotExist(err) {
						fmt.Printf("\tdeleted:    %s\n", path)
					} else {
						fmt.Printf("\tmodified:   %s\n", path)
					}
				}
			}

			// Print untracked files
			if len(untrackedFiles) > 0 {
				fmt.Println("\nUntracked files:")
				fmt.Println("  (use \"orb add <file>...\" to include in what will be committed)")

				var files []string
				for path := range untrackedFiles {
					files = append(files, path)
				}

				sort.Strings(files)
				for _, path := range files {
					fmt.Printf("\t%s\n", path)
				}

				fmt.Println("\nnothing added to commit but untracked files present (use \"orb add\" to track)")
			} else if len(modifiedFiles) == 0 && len(stagedFiles) == 0 {
				fmt.Println("nothing to commit, working tree clean")
			}

			return nil
		},
	}

	return cmd
}

// listWorkingDirFiles returns a list of all files in the working directory
// excluding the .orb directory and other files that would be ignored by Git
func listWorkingDirFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip .orb directory and hidden files/directories
		if strings.HasPrefix(path, ".orb/") ||
			(strings.HasPrefix(filepath.Base(path), ".") && path != ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories, we only want files
		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// isFileModified checks if a file has been modified since it was added to the index
func isFileModified(path string, entry index.Entry) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	// in real implementation, we compute the hash and compare
	// rn we can compare file size as a quick check
	if uint32(info.Size()) != entry.Size {
		return true, nil
	}

	// compare modification time
	// Note: In a real implementation, you would compute the hash and compare
	if info.ModTime().After(entry.ModTime) {
		return true, nil
	}

	return false, nil
}
