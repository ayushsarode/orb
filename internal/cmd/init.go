package cmd

import (
	"fmt"

	"github.com/ayushsarode/orb/internal/filesystem"
	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "init",
        Short: "Initialize a new orb repository",
        RunE: func(cmd *cobra.Command, args []string) error {
            fmt.Println("Starting repository initialization...")
            err := filesystem.InitRepository()
            if err != nil {
                fmt.Printf("Initialization failed: %v\n", err)
                return err
            }
            fmt.Println("Repository initialization completed successfully.")
            return nil
        },
    }
    return cmd
}