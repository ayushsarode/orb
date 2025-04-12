package cmd

import (
	"github.com/ayushsarode/orb/internal/filesystem"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new orb repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return filesystem.InitRepository()
		},
	}
	return cmd
}
