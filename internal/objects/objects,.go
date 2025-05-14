package objects

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const objectsDir = ".orb/objects"

const (
	BlobType   = "blob"
	TreeType   = "tree"
	CommitType = "commit"
)

func WriteObject(objType string, content []byte) (string, error) {
	// <type> of Object, <size> of content
	header := fmt.Sprintf("%s %d\x00", objType, len(content))

	data := append([]byte(header), content...)

	//calculate SHA-1 hash
	hash := sha1.Sum(data)

	hashStr := fmt.Sprintf("%x", hash)

	// stores objects in subdir based on their hash, first 2 char as a dir name
	objDir := filepath.Join(objectsDir, hashStr[:2])
	if err := os.MkdirAll(objDir, 0755); err != nil {
		return "", fmt.Errorf("creating object directory: %w ", err)
	}

	// write the file and compress
	objPath := filepath.Join(objDir, hashStr[2:])
	file, err := os.Create(objPath)

	if err != nil {
		return "", fmt.Errorf("creating object file: %w", err)
	}
	defer file.Close()

	zw := zlib.NewWriter(file)
	if _, err := zw.Write(data); err != nil {
		return "", fmt.Errorf("writing compressed  data: %w", err)
	}

	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("closing zlib writer: %w", err)
	}

	return hashStr, nil
}

func ReadObject(hash string) (string, []byte, error) {
	if len(hash) < 3 {
		return "", nil, fmt.Errorf("invalid hash: %s", hash)
	}

	// builds the file path and opens the compressed object file
	objPath := filepath.Join(objectsDir, hash[:2], hash[2:])
	file, err := os.Open(objPath)
	if err != nil {
		return "", nil, fmt.Errorf("opening object file: %w", err)
	}
	defer file.Close()

	zr, err := zlib.NewReader(file)
	if err != nil {
		return "", nil, fmt.Errorf("Creating zlib reader: %w", err)
	}
	defer zr.Close()

	data, err := io.ReadAll(zr)
	if err != nil {
		return "", nil, fmt.Errorf("reading object data: %w", err)
	}

	// splits the data into 2 parts
	// Header: "blob 12" and Content: "hello world\n"
	parts := bytes.SplitN(data, []byte{0}, 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid object format")
	}

	headerParts := bytes.SplitN(parts[0], []byte{' '}, 2)
	if len(headerParts) != 2 {
		return "", nil, fmt.Errorf("invalid object header")
	}

	// convert the object type to string, this will gets the actual content of the blob
	objType := string(headerParts[0])
	content := parts[1]

	return objType, content, nil
}

func WriteBlob(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)

	if err != nil {
		return "", fmt.Errorf("reading file: %w", err)
	}
	return WriteObject(BlobType, content)
}

// FindMissingObjects identifies objects needed from a remote repository
func FindMissingObjects(wantHash string, haveHash string) ([]string, error) {
	// In a full implementation, this would:
	// 1. Traverse the commit graph from wantHash
	// 2. Compare with objects reachable from haveHash
	// 3. Return a list of hashes that need to be fetched

	// For simplicity in this version, we'll just return wantHash
	// if it doesn't exist locally
	if haveHash == wantHash {
		return []string{}, nil
	}

	// Check if we already have the object
	objPath := filepath.Join(objectsDir, wantHash[:2], wantHash[2:])
	if _, err := os.Stat(objPath); err == nil {
		// Object exists locally
		return []string{}, nil
	}

	// We don't have this object
	return []string{wantHash}, nil
}

// ProcessPackData handles the packfile data received from a remote
func ProcessPackData(data []byte) (int, error) {
	// In a full implementation, this would:
	// 1. Parse the packfile format
	// 2. Extract and store the contained objects
	// 3. Return the count of processed objects

	// For now, we'll just simulate storing a single object
	// This is a placeholder that should be replaced with proper pack file processing

	// Simple validation that we received something
	if len(data) == 0 {
		return 0, fmt.Errorf("empty pack data received")
	}

	// In a real implementation, we would parse the packfile and extract objects
	fmt.Println("Pack data received, length:", len(data))

	// Return 1 to indicate we processed an object (placeholder)
	return 1, nil
}

// CollectCommitObjects gathers all objects related to a commit for pushing
func CollectCommitObjects(commitHash string) (map[string]struct {
	Type string
	Data []byte
}, error) {
	// In a full implementation, this would:
	// 1. Traverse the commit and its tree
	// 2. Collect all related objects (blobs, trees, commits)

	// For now, return just the commit itself
	objects := make(map[string]struct {
		Type string
		Data []byte
	})

	// Read the commit object
	objType, content, err := ReadObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("reading commit object: %w", err)
	}

	// Add it to our collection
	objects[commitHash] = struct {
		Type string
		Data []byte
	}{
		Type: objType,
		Data: content,
	}

	fmt.Println("Collected commit object:", commitHash)

	return objects, nil
}

// CheckoutCommit updates the working directory to match a commit
func CheckoutCommit(commitHash string) error {
	// In a full implementation, this would:
	// 1. Read the commit to get its tree
	// 2. Read the tree recursively
	// 3. Update the working directory to match the tree

	// Read the commit object to get the tree hash
	objType, content, err := ReadObject(commitHash)
	if err != nil {
		return fmt.Errorf("reading commit object: %w", err)
	}

	if objType != CommitType {
		return fmt.Errorf("expected commit object, got %s", objType)
	}

	// Parse the commit content to get the tree hash
	lines := strings.Split(string(content), "\n")
	var treeHash string

	for _, line := range lines {
		if strings.HasPrefix(line, "tree ") {
			treeHash = strings.TrimPrefix(line, "tree ")
			break
		}
	}

	if treeHash == "" {
		return fmt.Errorf("no tree found in commit")
	}

	fmt.Printf("Found tree %s in commit %s\n", treeHash, commitHash)
	fmt.Println("Checkout complete (placeholder)")

	// In a real implementation, we would update the working directory here

	return nil
}
