package cmd

import (
	"fmt"

	"github.com/ayushsarode/orb/internal/config"
	"github.com/ayushsarode/orb/internal/objects"
	"github.com/ayushsarode/orb/internal/refs"
	"github.com/ayushsarode/orb/internal/transport"
	"github.com/spf13/cobra"
)

func newPushCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push [remote] [branch]",
		Short: "Push commits to the remote repository",
		Long:  `Upload local branch commits to OrbHub or other remote repository`,
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

			// Check if the branch exists
			branchRef := fmt.Sprintf("refs/heads/%s", branchName)
			commitHash, err := refs.ReadRef(branchRef)
			if err != nil {
				return fmt.Errorf("branch %s does not exist: %w", branchName, err)
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

			fmt.Printf("Pushing to %s (%s)\n", remoteName, remote.URL)

			// Collect objects that need to be pushed
			fmt.Println("Collecting objects to push...")

			objectsToPush, err := objects.CollectCommitObjects(commitHash)
			if err != nil {
				return fmt.Errorf("failed to collect objects: %w", err)
			}

			fmt.Printf("Found %d objects to push\n", len(objectsToPush))

			// Push each object
			for hash, obj := range objectsToPush {
				fmt.Printf("Pushing object %s (%s)\n", hash[:8], obj.Type)
				err := remote.PushObject(obj.Type, hash, obj.Data)
				if err != nil {
					return fmt.Errorf("failed to push object %s: %w", hash, err)
				}
			}

			// Update the remote ref
			err = remote.UpdateRef(fmt.Sprintf("refs/heads/%s", branchName), commitHash)
			if err != nil {
				return fmt.Errorf("failed to update remote ref: %w", err)
			}

			fmt.Printf("\nSuccess! Pushed to %s/%s\n", remoteName, branchName)

			return nil
		},
	}

	return cmd
}
