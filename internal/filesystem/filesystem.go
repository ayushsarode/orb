// internal/filesystem/filesystem.go
package filesystem

import (
	"fmt"
	"os"
)

// Repository directories
const (
	OrbDir       = ".orb"
	ObjectsDir   = ".orb/objects"
	RefsDir      = ".orb/refs"
	RefsHeadsDir = ".orb/refs/heads"
	RefsTagsDir  = ".orb/refs/tags"
)

// Repository files
const (
	IndexFile  = ".orb/index"
	HeadFile   = ".orb/HEAD"
	ConfigFile = ".orb/config"
)

// InitRepository creates the basic directory structure for a new orb repository
func InitRepository() error {
	dirs := []string{
		OrbDir,
		ObjectsDir,
		RefsDir,
		RefsHeadsDir,
		RefsTagsDir,
	}

	// Create directories
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
		
	}

	// Create HEAD file pointing to main branch
	headContent := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(HeadFile, headContent, 0644); err != nil {
		return fmt.Errorf("creating HEAD file: %w", err)
	}
	
	// Create empty index file
	if _, err := os.Create(IndexFile); err != nil {
		return fmt.Errorf("creating index file: %w", err)
	}
	

	// Create basic config file
	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
`
	if err := os.WriteFile(ConfigFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	
	return nil
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
