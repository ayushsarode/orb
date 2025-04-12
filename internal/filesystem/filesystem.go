package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	OrbDir	= ".orb"
	ObjectsDir = ".orb/objects"
	RefsDir = ".orb/refs"
	RefsHeadsDir = ".orb/refs/heads"
	RefsTagsDir = ".orb/refs/tags"
)

const (
	Indexfile = ".orb/index"
	HeadFile = ".orb/HEAD"
	ConfigFile = ".orb/config"
)

func InitRepository() error {
	dirs :=  []string {
		OrbDir,
		ObjectsDir,
		RefsDir,
		RefsHeadsDir,
		RefsTagsDir,
	}

	for _,dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("Creating directory %s: %w", dir, err)
		}
	}

	// creates a .orb/HEAD file with content pointing to the default branch (typically main)
	if err := os.WriteFile(HeadFile, []byte("refs: refs/heads/main\n"), 0644); err != nil {
		return fmt.Errorf("creating HEAD file: %w", err)
	}

	// creates empty index file
	if _, err := os.Create(Indexfile); err != nil {
		return fmt.Errorf("creating index file: %w", err)
	}

	configContent := `[core]
	repositoryformatversion = 0
	filemode = true
	`

	if err := os.WriteFile(ConfigFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}


	fmt.Println("Initialized empty orb reposiporty in", filepath.Join(getCurrentDir(), OrbDir)) 
		return nil 
	}



	func getCurrentDir() string {
		dir, err := os.Getwd()
		if err != nil {
			return "."
		}

	return dir
}