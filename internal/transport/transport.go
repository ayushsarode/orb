package transport

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RemoteConfig represents the configuration for a remote repository
type RemoteConfig struct {
	Name string
	URL  string
}

// Remote represents a remote repository
type Remote struct {
	Name     string
	URL      string
	Client   *http.Client
	Username string
	Password string
}

// NewRemote creates a new remote with the given name and URL
func NewRemote(name, url string) *Remote {
	return &Remote{
		Name: name,
		URL:  url,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetAuth sets authentication for the remote
func (r *Remote) SetAuth(username, password string) {
	r.Username = username
	r.Password = password
}

// FetchRefs fetches remote refs like branches and tags
func (r *Remote) FetchRefs() (map[string]string, error) {
	endpoint := fmt.Sprintf("%s/info/refs?service=git-upload-pack", r.URL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "orb/1.0")
	req.Header.Set("Content-Type", "application/x-git-upload-pack-request")

	// Set auth if provided
	if r.Username != "" && r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server responded with status: %s", resp.Status)
	}

	// Parse the response
	refs := make(map[string]string)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Skip the service header lines
	lines := strings.Split(string(body), "\n")
	startIdx := 0
	for i, line := range lines {
		if line == "" {
			startIdx = i + 1
			break
		}
	}

	// Parse ref lines - format is: <sha>\t<ref-name>
	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		if len(line) == 0 {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) != 2 {
			continue
		}

		// Remove the leading zero length and cap^ if present
		sha := strings.TrimPrefix(parts[0], "0000")
		sha = strings.TrimSuffix(sha, "^{}")
		refs[parts[1]] = sha
	}

	return refs, nil
}

// PushObject sends an object to the remote repository
func (r *Remote) PushObject(objectType string, hash string, data []byte) error {
	endpoint := fmt.Sprintf("%s/git-receive-pack", r.URL)

	// Prepare the request body with Git pack format
	// This is a simplified version, actual implementation would need proper pack format
	body := bytes.NewBuffer(data)

	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "orb/1.0")
	req.Header.Set("Content-Type", "application/x-git-receive-pack-request")

	if r.Username != "" && r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server rejected push with status: %s", resp.Status)
	}

	return nil
}

// UpdateRef updates a reference on the remote repository
func (r *Remote) UpdateRef(refName, targetHash string) error {
	endpoint := fmt.Sprintf("%s/git-receive-pack", r.URL)

	// Format the update reference request according to Git protocol
	var requestBody bytes.Buffer
	fmt.Fprintf(&requestBody, "%s %s refs/heads/\x00 report-status\n", targetHash, refName)
	fmt.Fprintf(&requestBody, "0000") // Flush packet

	req, err := http.NewRequest("POST", endpoint, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "orb/1.0")
	req.Header.Set("Content-Type", "application/x-git-receive-pack-request")

	if r.Username != "" && r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update ref on remote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server rejected ref update with status: %s", resp.Status)
	}

	// Check response for success
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	response := string(respData)
	if strings.Contains(response, "error") || strings.Contains(response, "failure") {
		return fmt.Errorf("server rejected ref update: %s", response)
	}

	return nil
}

// CloneRepo initializes a new repository with objects from remote
func (r *Remote) CloneRepo(destPath string) error {
	// Implementation would:
	// 1. Get remote refs
	// 2. Fetch objects
	// 3. Initialize local repo
	// 4. Store objects
	// 5. Set up refs
	return errors.New("not implemented yet")
}

// FetchObjects fetches objects from remote repository
func (r *Remote) FetchObjects(wants []string, haves []string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/git-upload-pack", r.URL)

	// Format wants and haves according to Git protocol
	var requestBody bytes.Buffer

	// Add wants
	for _, want := range wants {
		fmt.Fprintf(&requestBody, "want %s\n", want)
	}

	// Add haves if any
	if len(haves) > 0 {
		fmt.Fprint(&requestBody, "\n")
		for _, have := range haves {
			fmt.Fprintf(&requestBody, "have %s\n", have)
		}
		fmt.Fprint(&requestBody, "done\n")
	} else {
		fmt.Fprint(&requestBody, "\ndone\n")
	}

	req, err := http.NewRequest("POST", endpoint, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "orb/1.0")
	req.Header.Set("Content-Type", "application/x-git-upload-pack-request")

	if r.Username != "" && r.Password != "" {
		req.SetBasicAuth(r.Username, r.Password)
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from remote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server rejected fetch with status: %s", resp.Status)
	}

	packData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return packData, nil
}

// IsValidURL checks if a given string is a valid remote URL
func IsValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}
