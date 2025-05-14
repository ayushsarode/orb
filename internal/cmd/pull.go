package cmd

import (
	"fmt"

	"github.com/ayushsarode/orb/internal/config"
	"github.com/ayushsarode/orb/internal/objects"
	"github.com/ayushsarode/orb/internal/refs"
	"github.com/ayushsarode/orb/internal/transport"
	"github.com/spf13/cobra"
)

func newPullCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull [remote] [branch]",
		Short: "Pull changes from a remote repository",
		Long:  `Pull latest changes from OrbHub or other remote repository and merge them into current branch`,
		Args:  cobra.RangeArgs(0, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default remote and branch
			remoteName := "origin"
			branchName := "main"

			// Override with user-provided values if available
			if len(args) >= 1 {
				remoteName = args[0]
			}
			if len(args) >= 2 {
				branchName = args[1]
			}

			// Load the remote configuration
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Find the specified remote
			var remoteConfig transport.RemoteConfig
			found := false
			for _, r := range cfg.Remotes {
				if r.Name == remoteName {
					remoteConfig = r
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("remote '%s' not found, use 'orb remote add' to add it first", remoteName)
			}

			// Create a remote instance
			remote := transport.NewRemote(remoteConfig.Name, remoteConfig.URL)

			// Check for credentials
			username := cfg.GetString(fmt.Sprintf("remote.%s.username", remoteName), "")
			password := cfg.GetString(fmt.Sprintf("remote.%s.password", remoteName), "")
			if username != "" && password != "" {
				remote.SetAuth(username, password)
			}

			fmt.Printf("Pulling from %s (%s)\n", remoteName, remote.URL)

			// Fetch remote refs
			remoteRefs, err := remote.FetchRefs()
			if err != nil {
				return fmt.Errorf("failed to fetch refs from remote: %w", err)
			}

			remoteRefName := fmt.Sprintf("refs/heads/%s", branchName)
			remoteCommitHash, ok := remoteRefs[remoteRefName]
			if !ok {
				return fmt.Errorf("branch '%s' not found on remote '%s'", branchName, remoteName)
			}

			fmt.Printf("Remote branch '%s' points to commit %s\n", branchName, remoteCommitHash[:8])

			// Get our current HEAD
			headRef, err := refs.ReadHead()
			if err != nil {
				return fmt.Errorf("failed to read HEAD: %w", err)
			}

			// Get our current branch commit
			localCommitHash, err := refs.ReadRef(headRef)
			if err != nil {
				return fmt.Errorf("failed to read current branch: %w", err)
			}

			fmt.Printf("Local branch '%s' points to commit %s\n", headRef, localCommitHash[:8])

			if remoteCommitHash == localCommitHash {
				fmt.Println("Already up to date!")
				return nil
			}

			// Determine what objects we need to fetch
			missingObjects, err := objects.FindMissingObjects(remoteCommitHash, localCommitHash)
			if err != nil {
				return fmt.Errorf("failed to find missing objects: %w", err)
			}

			fmt.Printf("Fetching %d objects...\n", len(missingObjects))

			// Fetch missing objects
			if len(missingObjects) > 0 {
				packData, err := remote.FetchObjects(missingObjects, []string{localCommitHash})
				if err != nil {
					return fmt.Errorf("failed to fetch objects: %w", err)
				}

				// Process and store the pack data
				processedCount, err := objects.ProcessPackData(packData)
				if err != nil {
					return fmt.Errorf("failed to process pack data: %w", err)
				}

				fmt.Printf("Received %d objects\n", processedCount)
			}

			// Update our ref to the new commit
			currentBranch := refs.GetCurrentBranch()
			if currentBranch == branchName {
				err := refs.UpdateRef(fmt.Sprintf("refs/heads/%s", branchName), remoteCommitHash)
				if err != nil {
					return fmt.Errorf("failed to update local ref: %w", err)
				}
			} else {
				fmt.Printf("Warning: You're not on branch '%s'. Fetch completed but not merged.\n", branchName)
				fmt.Printf("Use 'orb checkout %s' to switch to this branch.\n", branchName)
				return nil
			}

			fmt.Println("Fast-forward merge successful!")
			fmt.Printf("Successfully pulled from %s/%s\n", remoteName, branchName)
			return nil
		},
	}

	return cmd
}
