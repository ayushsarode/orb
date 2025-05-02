package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ayushsarode/orb/internal/refs"
	"github.com/spf13/cobra"
)

func newCheckoutCommand() *cobra.Command {
	var createBranch bool

	cmd := &cobra.Command{
		Use:   "checkout [branch/commit]",
		Short: "Switch branches or restore working tree files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("you must specify a branch name or commit hash")
			}

			target := args[0]

			 // try to get the reference first; if it succeeds, it's a valid branch or tag
			_, err := refs.GetRef(target)
			branchExists := err == nil

			// If -b flag is specified, create a new branch
			if createBranch {
				if branchExists {
					return fmt.Errorf("branch '%s' already exists", target)
				}

				// Get current HEAD commit
				head, err := refs.GetHead()
				if err != nil {
					return fmt.Errorf("getting HEAD: %w", err)
				}

				// create new branch + target is the name of branch added to refs/heads with current HEAD
				if err := refs.UpdateRef("refs/heads/"+target, head); err != nil {
					return fmt.Errorf("creating branch: %w", err)
				}

				fmt.Printf("Created branch '%s'\n", target)
				branchExists = true
			}

			// If it's a branch, update HEAD to point to it
			if branchExists {
				if err := refs.UpdateHead(target); err != nil {
					return fmt.Errorf("switching to branch: %w", err)
				}
				fmt.Printf("Switched to branch '%s'\n", target)
				return nil
			}

			// If it's not a branch, check if it's a commit hash
			if len(target) == 40 {
				// Verify this is a valid commit
				hashPath := filepath.Join(".orb/objects", target[:2], target[2:])
				if _, err := os.Stat(hashPath); err != nil {
					return fmt.Errorf("commit '%s' does not exist", target)
				}

				// Set HEAD to point directly to the commit (detached HEAD)
				if err := refs.UpdateHead(target); err != nil {
					return fmt.Errorf("checking out commit: %w", err)
				}

				fmt.Printf("Note: you are in 'detached HEAD' state at %s\n", target[:7])
				return nil
			}

			return fmt.Errorf("'%s' is not a branch or commit hash", target)
		},
	}

	cmd.Flags().BoolVarP(&createBranch, "create-branch", "b", false, "Create and checkout a new branch")

	return cmd
}
