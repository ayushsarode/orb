package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ayushsarode/orb/internal/refs"
	"github.com/spf13/cobra"
)

func newBranchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "branch [branchname]",
		Short: "List, create, or delete branches",
		RunE: func(cmd *cobra.Command, args []string) error {

			// if its just "orb branch" then return list of branches
			if len(args) == 0 {
				return listBranches()
			}

			// create a new branch
			branchName := args[0]

			// get current HEAD commit
			head, err := refs.GetHead()
			if err != nil {
				return fmt.Errorf("getting HEAD: %w", err)
			}

			// create the branch (a ref pointing to the current commit)
			if err := refs.UpdateRef("refs/heads/"+branchName, head); err != nil {
				return fmt.Errorf("creating branch: %w", err)
			}

			fmt.Printf("Created branch '%s'\n", branchName)
			return nil
		},
	}

	return cmd
}

func listBranches() error {
	// get current branch
	currentBranch := ""
	headContent, err := os.ReadFile(".orb/HEAD")
	if err == nil {
		head := strings.TrimSpace(string(headContent))
		if strings.HasPrefix(head, "ref: refs/heads/") {
			currentBranch = strings.TrimPrefix(head, "ref: refs/heads/")
		}
	}

	// read branch directory
	branches, err := os.ReadDir(".orb/refs/heads")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading branches directory: %w", err)
	}

	// listing branches
	for _, branch := range branches {
		if branch.IsDir() {
			continue 
		}

		branchName := branch.Name()
		if branchName == currentBranch {
			fmt.Printf("* %s\n", branchName)
		} else {
			fmt.Printf("  %s\n", branchName)
		}
	}

	return nil
}
