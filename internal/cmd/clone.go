package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ayushsarode/orb/internal/config"
	"github.com/ayushsarode/orb/internal/filesystem"
	"github.com/ayushsarode/orb/internal/objects"
	"github.com/ayushsarode/orb/internal/refs"
	"github.com/ayushsarode/orb/internal/transport"
	"github.com/spf13/cobra"
)

func newCloneCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone <url> [directory]",
		Short: "Clone a remote repository",
		Long:  `Clone a repository from OrbHub or other remote location into a new directory`,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			// Check if URL format is valid
			if !transport.IsValidURL(url) {
				return fmt.Errorf("invalid URL format: %s - must begin with http:// or https://", url)
			}

			// Determine the destination directory
			var destDir string
			if len(args) > 1 {
				destDir = args[1]
			} else {
				// Extract repo name from URL for default directory name
				parts := strings.Split(url, "/")
				repoName := parts[len(parts)-1]
				repoName = strings.TrimSuffix(repoName, ".git")
				destDir = repoName
			}

			// Check if destination directory already exists
			if _, err := os.Stat(destDir); !os.IsNotExist(err) {
				return fmt.Errorf("destination path '%s' already exists and is not an empty directory", destDir)
			}

			fmt.Printf("Cloning into '%s'...\n", destDir)

			// Create the destination directory
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory '%s': %w", destDir, err)
			}

			// Switch to the destination directory
			originalWorkDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			if err := os.Chdir(destDir); err != nil {
				return fmt.Errorf("failed to change directory to '%s': %w", destDir, err)
			}

			// Initialize a new repository
			if err := filesystem.InitRepository(); err != nil {
				// Change back to original directory before returning error
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to initialize repository: %w", err)
			}

			// Set up the remote
			remoteName := "origin"
			remote := transport.NewRemote(remoteName, url)

			// Fetch remote refs
			fmt.Println("Fetching remote refs...")
			remoteRefs, err := remote.FetchRefs()
			if err != nil {
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to fetch refs from remote: %w", err)
			}

			// Look for the main ref (main or master)
			mainRef := "refs/heads/main"
			mainCommit, ok := remoteRefs[mainRef]
			if !ok {
				// Try master branch as fallback
				mainRef = "refs/heads/master"
				mainCommit, ok = remoteRefs[mainRef]
				if !ok {
					_ = os.Chdir(originalWorkDir)
					return fmt.Errorf("remote repository has no 'main' or 'master' branch")
				}
			}

			// Set the default branch name based on what we found
			defaultBranch := strings.TrimPrefix(mainRef, "refs/heads/")

			fmt.Printf("Remote %s branch found at commit %s\n", defaultBranch, mainCommit[:8])

			// Fetch objects
			fmt.Println("Fetching objects...")
			wants := []string{mainCommit}
			haves := []string{} // We have no objects yet

			packData, err := remote.FetchObjects(wants, haves)
			if err != nil {
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to fetch objects: %w", err)
			}

			// Process and store the fetched objects
			count, err := objects.ProcessPackData(packData)
			if err != nil {
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to process pack data: %w", err)
			}

			fmt.Printf("Received %d objects\n", count)

			// Update refs to point to the fetched commit
			localBranchName := defaultBranch
			localRef := fmt.Sprintf("refs/heads/%s", localBranchName)

			if err := refs.UpdateRef(localRef, mainCommit); err != nil {
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to update ref '%s': %w", localRef, err)
			}

			// Point HEAD to our default branch
			if err := refs.UpdateHead(localRef); err != nil {
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to update HEAD: %w", err)
			}

			// Update configuration to remember the remote
			cfg, err := config.LoadConfig()
			if err != nil {
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to load config: %w", err)
			}

			cfg.Remotes = append(cfg.Remotes, transport.RemoteConfig{
				Name: remoteName,
				URL:  url,
			})

			if err := config.SaveConfig(cfg); err != nil {
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to save config: %w", err)
			}

			// Set the upstream tracking branch
			branchConfigKey := fmt.Sprintf("branch.%s.remote", localBranchName)
			cfg.Set(branchConfigKey, remoteName)

			mergeConfigKey := fmt.Sprintf("branch.%s.merge", localBranchName)
			cfg.Set(mergeConfigKey, fmt.Sprintf("refs/heads/%s", defaultBranch))

			if err := cfg.Save(); err != nil {
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to save branch tracking config: %w", err)
			}

			// Checkout the working directory
			if err := objects.CheckoutCommit(mainCommit); err != nil {
				_ = os.Chdir(originalWorkDir)
				return fmt.Errorf("failed to checkout files: %w", err)
			}

			// Change back to original directory
			if err := os.Chdir(originalWorkDir); err != nil {
				fmt.Printf("Warning: Failed to change back to original directory: %v\n", err)
			}

			fmt.Println("\nClone completed successfully!")
			fmt.Printf("Repository cloned into '%s'\n", destDir)

			return nil
		},
	}

	return cmd
}
