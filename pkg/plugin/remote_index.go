package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RemoteIndex fetches a plugin marketplace index from a remote HTTP endpoint.
// This is a read-only PoC: it fetches a JSON array of PluginMeta from a URL
// and merges results with the local index for search.
type RemoteIndex struct {
	url    string
	client *http.Client
}

// NewRemoteIndex creates a RemoteIndex pointing at the given URL.
func NewRemoteIndex(url string) *RemoteIndex {
	return &RemoteIndex{
		url:    url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Fetch retrieves the remote plugin list. Returns empty list on error
// (fail-open: remote unavailability does not block local search).
func (r *RemoteIndex) Fetch() ([]PluginMeta, error) {
	if r.url == "" {
		return nil, nil
	}
	resp, err := r.client.Get(r.url)
	if err != nil {
		return nil, fmt.Errorf("fetch remote index: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remote index returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read remote index body: %w", err)
	}
	var plugins []PluginMeta
	if err := json.Unmarshal(body, &plugins); err != nil {
		return nil, fmt.Errorf("parse remote index: %w", err)
	}
	return plugins, nil
}

// SearchMerged combines local and remote plugin lists, then filters by query.
func SearchMerged(local *Marketplace, remote *RemoteIndex, query string) ([]PluginMeta, error) {
	localPlugins, err := local.Search(query)
	if err != nil {
		return nil, err
	}
	remotePlugins, err := remote.Fetch()
	if err != nil {
		// fail-open: remote errors do not block local results
		return localPlugins, nil
	}
	q := lowercase(query)
	var results []PluginMeta
	for _, p := range remotePlugins {
		if contains(lowercase(p.Name), q) ||
			contains(lowercase(p.Backend), q) ||
			contains(lowercase(p.Description), q) {
			results = append(results, p)
		}
	}
	// Deduplicate by name, local takes precedence
	seen := make(map[string]bool)
	for _, p := range localPlugins {
		seen[p.Name] = true
	}
	for _, p := range results {
		if !seen[p.Name] {
			localPlugins = append(localPlugins, p)
		}
	}
	return localPlugins, nil
}

func lowercase(s string) string { return strings.ToLower(s) }
func contains(s, substr string) bool { return strings.Contains(s, substr) }