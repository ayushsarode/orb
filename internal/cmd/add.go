package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"github.com/ayushsarode/orb/internal/index"
	"github.com/ayushsarode/orb/internal/objects"
	"github.com/spf13/cobra"
)

func newAddCommnad() *cobra.Command {
	cmd := &cobra.Command{
		Use: "add [files...]",
		Short: "Add file contents to the index",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("nothing specified, nothing added")
			}
			
			idx, err := index.LoadIndex()
			if err != nil {
				return fmt.Errorf("loading index: %w", err)
			}
			
			for _, pattern := range args {
			
				matches, err := filepath.Glob(pattern)
				if err != nil {
					return fmt.Errorf("invalid pattern: %w", err)
				}
				
				for _, path := range matches {
					
					info, err := os.Stat(path)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: cannot stat '%s': %v\n", path, err)
						continue
					}
					
					if info.IsDir() {
						// here we need to handle dir recursively
						fmt.Printf("Adding files in '%s' is not yet supported\n", path)
						continue
					}
					
					
					hash, err := objects.WriteBlob(path)
					if err != nil {
						return fmt.Errorf("writing blob for %s: %w", path, err)
					}
					
					// Add to the index
					if err := idx.AddFile(path, hash); err != nil {
						return fmt.Errorf("adding to index: %w", err)
					}
					
					fmt.Printf("Added '%s'\n", path)
				}
			}
			
			// write the updated index
			if err := idx.Write(); err != nil {
				return fmt.Errorf("writing index: %w", err)
			}
			
			return nil
		},
	}
	return cmd
}