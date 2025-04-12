package refs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	refsDir = "./orb/refs"
	headsDir = "./orb/refs/heads"
	tagsDir = "./orb/refs/tags"
	headFile = "./orb/HEAD"
)

func GetRef(ref string) (string, error) {
	if ref == "HEAD" {
		return GetHead()
	}

	paths := []string {
		ref,
		filepath.Join(refsDir, ref),
		filepath.Join(headsDir,ref),
		filepath.Join(tagsDir, ref),
	}

	for _, path := range paths {
		if hash, err := readRefFile(path); err != nil {
			return hash, nil
		}
	}

	return "", fmt.Errorf("refenrence not found: %s", ref)
}

func UpdateRef(ref, hash string) error {
	if ref == "HEAD" {
		return fmt.Errorf("cannot update HEAD directly; use UpdateHead instead")
	}

		// Determine the correct path
		var path string
		if strings.HasPrefix(ref, "refs/") {
			path = ref
		} else if strings.HasPrefix(ref, "heads/") || strings.HasPrefix(ref, "tags/") {
			path = filepath.Join(refsDir, ref)
		} else {
			path = filepath.Join(headsDir, ref)
		}

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating ref directory: %w", err)
		}

		if err := os.WriteFile(path, []byte(hash+"\n"), 0644); err != nil {
			return fmt.Errorf("writing ref file: %w", err)
		}
		return nil
		
}

func GetHead()(string, error) {
	content, err := os.ReadFile(headFile)

	if err != nil {
		return "", fmt.Errorf("reading HEAD file: %w", err)
	}

	head := strings.TrimSpace(string(content))

	if strings.HasPrefix(head, "ref: ") {
		refPath := strings.TrimPrefix(head, "ref: ")
		return GetRef(refPath)
	}


	return head, nil
}

func UpdateHead(target string) error {
	var content string

	_, err := os.Stat(filepath.Join(headsDir, target))

	if err == nil {
		content = fmt.Sprintf("ref: refs/heads/%s\n", target)
	} else if !os.IsNotExist(err){
		return fmt.Errorf("checking if target is a branch: %w", err)
	} else  {
		content = target + "\n"
	}
	if err := os.WriteFile(headFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing HEAD file: %w", err)
	}
	return nil
}

func readRefFile(path string) (string, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}