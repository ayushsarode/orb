package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCommnad() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "orb",
		Short: "orb is a version control system",
		Long:  `orb is a distributed version control,system inspired by Git but written in Go.`,
	}
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newAddCommnad())
	rootCmd.AddCommand(newCommitCommand())
	rootCmd.AddCommand(newLogCommand())
	rootCmd.AddCommand(newStatusCommand())
	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newBranchCommand())
	rootCmd.AddCommand(newCheckoutCommand())

	// Add remote repository commands
	rootCmd.AddCommand(newRemoteCommand())
	// rootCmd.AddCommand(newPushCommand())
	// rootCmd.AddCommand(newPullCommand())
	rootCmd.AddCommand(newCloneCommand())

	return rootCmd
}
