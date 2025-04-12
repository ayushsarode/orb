package cmd

import (
"github.com/spf13/cobra"
)

func NewRootCommnad() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "orb",
		Short: "orb is a version control system",
		Long:  `orb is a distributed version control,system inspired by Git but written in Go.`,

	}
	rootCmd.AddCommand(newInitCommand())
	rootCmd.AddCommand(newAddCommnad())
	// rootCmd.AddCommand(NewCommitCommnad())
	// rootCmd.AddCommand(newLogComnad())


	return rootCmd
}