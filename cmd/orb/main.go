package main

import (
	"fmt"
	"os"
	"github.com/ayushsarode/orb/internal/cmd"
)

func main() {
	rootCmd := cmd.NewRootCommnad()
	if err := rootCmd.Execute(); err != nil {
		 fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		 os.Exit(1)
	}
}