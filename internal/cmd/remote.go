package cmd

import (
	"errors"
	"fmt"

	"github.com/ayushsarode/orb/internal/config"
	"github.com/ayushsarode/orb/internal/transport"
	"github.com/spf13/cobra"
)

func newRemoteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage remote repositories",
		Long:  `Add, remove, or show remote repositories that can sync with OrbHub`,
	}

	// Add subcommands
	cmd.AddCommand(newRemoteAddCommand())
	cmd.AddCommand(newRemoteRemoveCommand())
	cmd.AddCommand(newRemoteListCommand())

	return cmd
}

func newRemoteAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a new remote repository",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			url := args[1]

			if !transport.IsValidURL(url) {
				return errors.New("invalid URL format - must begin with http:// or https://")
			}

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Check if remote already exists
			for _, remote := range cfg.Remotes {
				if remote.Name == name {
					return fmt.Errorf("remote %s already exists", name)
				}
			}

			// Add new remote
			cfg.Remotes = append(cfg.Remotes, transport.RemoteConfig{
				Name: name,
				URL:  url,
			})

			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Remote '%s' added with URL '%s'\n", name, url)
			return nil
		},
	}
	return cmd
}

func newRemoteRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a remote repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			found := false
			newRemotes := []transport.RemoteConfig{}

			for _, remote := range cfg.Remotes {
				if remote.Name != name {
					newRemotes = append(newRemotes, remote)
				} else {
					found = true
				}
			}

			if !found {
				return fmt.Errorf("remote '%s' not found", name)
			}

			cfg.Remotes = newRemotes

			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Remote '%s' removed\n", name)
			return nil
		},
	}
	return cmd
}

func newRemoteListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List remote repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if len(cfg.Remotes) == 0 {
				fmt.Println("No remotes configured")
				return nil
			}

			fmt.Println("Configured remotes:")
			for _, remote := range cfg.Remotes {
				fmt.Printf("%s\t%s\n", remote.Name, remote.URL)
			}
			return nil
		},
	}
	return cmd
}
