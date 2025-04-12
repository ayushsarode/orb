package objects

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const objectsDir = ".orb/objects"

const (
	BlobType = "blob"
	TreeType = "tree"
	CommitType = "commit"
)

func WriteObject(objType string, content []byte)(string, error) {


	// <type> of Object, <size> of content
	header := fmt.Sprintf("%s %d\x00", objType, len(content))

	data := append([]byte(header), content...)


	//calculate SHA-1 hash
	hash := sha1.Sum(data)

	hashStr  := fmt.Sprintf("%x", hash)

	// stores objects in subdir based on their hash, first 2 char as a dir name
	objDir := filepath.Join(objectsDir, hashStr[:2])
	if err := os.MkdirAll(objDir, 0755); err != nil {
		return "", fmt.Errorf("creating object directory: %w ", err)
	}

	// write the file and compress
	objPath := filepath.Join(objDir, hashStr[2:])
	file,err := os.Create(objPath)

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
		return "",nil, fmt.Errorf("invalid hash: %s", hash)

	}

	// builds the file path and opens the compressed object file
	objPath := filepath.Join(objectsDir, hash[:2], hash[2:])
	file, err := os.Open(objPath)
	if err != nil {
		return "", nil,fmt.Errorf("opening object file: %w", err)
	}
	defer file.Close()

	zr, err := zlib.NewReader(file)
	if err != nil {
		return "", nil, fmt.Errorf("Creating zlib reader: %w", err)
	}
	defer zr.Close()

	data,err := io.ReadAll(zr)
	if err !=nil {
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
		return "",nil, fmt.Errorf("invalid object header")
	}

	// convert the object type to string, this will gets the actual content of the blob
	objType := string(headerParts[0])
	content := parts[1]

	return objType, content, nil
}

func WriteBlob(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)

	if err !=nil {
		return "", fmt.Errorf("reading file: %w", err)
	}
	return WriteObject(BlobType, content)
}