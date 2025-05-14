package refs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// other pkgs needed to access these pathss which required exported (capitalized) const
const (
	RefsDir  = ".orb/refs"
	HeadsDir = ".orb/refs/heads" // where branches live
	TagsDir  = ".orb/refs/tags"  // where tags live
	HeadFile = ".orb/HEAD"       // special pointer to current location
)

// ReadRef reads the commit hash that a ref points to
func ReadRef(ref string) (string, error) {
	var path string

	// For refs with standard format
	if strings.HasPrefix(ref, "refs/") {
		path = filepath.Join(".orb", ref)
	} else if strings.HasPrefix(ref, "heads/") || strings.HasPrefix(ref, "tags/") {
		path = filepath.Join(RefsDir, ref)
	} else {
		// Assume it's a branch name
		path = filepath.Join(HeadsDir, ref)
	}

	return readRefFile(path)
}

// ReadHead reads the current HEAD reference
func ReadHead() (string, error) {
	content, err := os.ReadFile(HeadFile)
	if err != nil {
		return "", fmt.Errorf("reading HEAD file: %w", err)
	}

	head := strings.TrimSpace(string(content))

	// If HEAD points to a ref (normal case)
	if strings.HasPrefix(head, "ref: ") {
		return strings.TrimPrefix(head, "ref: "), nil
	}

	// Detached HEAD (points directly to a commit)
	return head, nil
}

// GetCurrentBranch returns the name of the current branch
func GetCurrentBranch() string {
	content, err := os.ReadFile(HeadFile)
	if err != nil {
		return ""
	}

	head := strings.TrimSpace(string(content))

	// If HEAD points to a branch
	if strings.HasPrefix(head, "ref: refs/heads/") {
		return strings.TrimPrefix(head, "ref: refs/heads/")
	}

	// Detached HEAD state
	return ""
}

func GetRef(ref string) (string, error) {
	// First check if this is a full ref path
	var path string

	if ref == "HEAD" {
		// Special case for HEAD - read the HEAD file directly first
		headContent, err := os.ReadFile(HeadFile)
		if err != nil {
			return "", fmt.Errorf("reading HEAD file: %w", err)
		}

		head := strings.TrimSpace(string(headContent))
		if strings.HasPrefix(head, "ref: ") {
			refPath := strings.TrimPrefix(head, "ref: ")
			return GetRef(refPath)
		}
		return head, nil
	} else if strings.HasPrefix(ref, "refs/") {
		path = filepath.Join(".orb", ref)
	} else if strings.HasPrefix(ref, "heads/") || strings.HasPrefix(ref, "tags/") {
		path = filepath.Join(RefsDir, ref)
	} else {
		// Try as a branch name first
		path = filepath.Join(HeadsDir, ref)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// If not a branch, try as a tag name
			path = filepath.Join(TagsDir, ref)
		}
	}

	return readRefFile(path)
}

// changes where a reference points
func UpdateRef(ref, hash string) error {
	if ref == "HEAD" {
		return fmt.Errorf("cannot update HEAD directly; use UpdateHead instead")
	}

	var path string

	// For branches: writes the commit hash to a file in heads
	if strings.HasPrefix(ref, "refs/") {
		if !strings.HasPrefix(ref, ".orb/") { //.orb/ prefix
			path = filepath.Join(".orb", ref)
		} else {
			path = ref
		}
		// For tags: writes the commit hash to a file in tags
	} else if strings.HasPrefix(ref, "heads/") || strings.HasPrefix(ref, "tags/") {
		path = filepath.Join(RefsDir, ref)
	} else {
		path = filepath.Join(HeadsDir, ref)
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

func GetHead() (string, error) {
	content, err := os.ReadFile(HeadFile)

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
	// If target has the format of a hash (40 hex characters),
	// we're in detached HEAD mode
	if len(target) == 40 && isValidHash(target) {
		content := target + "\n"
		if err := os.WriteFile(HeadFile, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing HEAD file: %w", err)
		}
		return nil
	}

	// Check if target exists as a branch
	branchPath := filepath.Join(HeadsDir, target)
	if _, err := os.Stat(branchPath); err != nil && os.IsNotExist(err) {
		return fmt.Errorf("checking branch existence: %w", err)
	}

	// Set HEAD to point to the branch
	content := fmt.Sprintf("ref: refs/heads/%s\n", target)
	if err := os.WriteFile(HeadFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing HEAD file: %w", err)
	}

	return nil
}

// check if a string is a valid hex hash
func isValidHash(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
func readRefFile(path string) (string, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}
