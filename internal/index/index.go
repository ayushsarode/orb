// internal/index/index.go
package index

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const indexFile = ".orb/index"

// entry represents a single entry in the index file
type Entry struct {
	Path       string
	ObjectHash string
	ModTime    time.Time
	Mode       uint32
	Size       uint32
}

// index represent the staging area 
// Index is a collection of all staged files
type Index struct {
	Entries map[string]Entry
}

// NewIndex creates a new empty index
func NewIndex() *Index {
	return &Index{
		Entries: make(map[string]Entry),
	}
}

// LoadIndex loads the index from disk
func LoadIndex() (*Index, error) {
	idx := NewIndex()
	
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return idx, nil
	}
	
	file, err := os.Open(indexFile)
	if err != nil {
		return idx, fmt.Errorf("opening index file: %w", err)
	}
	defer file.Close()
	
	// read index file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue 
		}
		
		hash := parts[0]
		path := parts[1]
		
		// cleaning up the path (remove any trailing %)
		path = strings.TrimSuffix(path, "%")
		
		// file info
		var modTime time.Time
		var mode, size uint32
		
		if info, err := os.Stat(path); err == nil {
			modTime = info.ModTime()
			mode = uint32(info.Mode())
			size = uint32(info.Size())
		}
		
		// Add to the index
		idx.Entries[path] = Entry{
			Path:       path,
			ObjectHash: hash,
			ModTime:    modTime,
			Mode:       mode,
			Size:       size,
		}
	}
	
	if err := scanner.Err(); err != nil {
		return idx, fmt.Errorf("reading index file: %w", err)
	}
	
	return idx, nil
}

// AddFile adds a file to the index means it stages a file for commit
func (idx *Index) AddFile(path string, hash string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("getting file info: %w", err)
	}
	
	idx.Entries[path] = Entry{
		Path:       path,
		ObjectHash: hash,
		ModTime:    info.ModTime(),
		Mode:       uint32(info.Mode()),
		Size:       uint32(info.Size()),
	}
	
	return nil
}

// Write writes the index to disk, basically saves all staged changes
func (idx *Index) Write() error {
	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(indexFile), 0755); err != nil {
		return fmt.Errorf("creating index directory: %w", err)
	}

	// Open the file for writing
	file, err := os.Create(indexFile)
	if err != nil {
		return fmt.Errorf("creating index file: %w", err)
	}
	defer file.Close()
	
	// sort entries by path for consistency
	var entries []Entry
	for _, entry := range idx.Entries {
		entries = append(entries, entry)
	}
	
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	
	// Write each entry
	for _, entry := range entries {
		line := fmt.Sprintf("%s %s\n", entry.ObjectHash, entry.Path)
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("writing entry to index: %w", err)
		}
	}
	
	return nil
}

// GetEntries returns all entries in the index, sorted by path
func (idx *Index) GetEntries() []Entry {
	var entries []Entry
	for _, entry := range idx.Entries {
		entries = append(entries, entry)
	}
	
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	
	return entries
}