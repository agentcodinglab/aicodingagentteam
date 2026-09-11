package plugin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoteIndex_EmptyURL_ReturnsNil(t *testing.T) {
	r := NewRemoteIndex("")
	plugins, err := r.Fetch()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if plugins != nil {
		t.Errorf("expected nil, got %v", plugins)
	}
}

func TestRemoteIndex_Fetch_OK(t *testing.T) {
	body := `[{"name":"test","backend":"test","version":"1.0.0"}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer srv.Close()

	r := NewRemoteIndex(srv.URL)
	plugins, err := r.Fetch()
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(plugins) != 1 || plugins[0].Name != "test" {
		t.Fatalf("expected 1 plugin test, got %v", plugins)
	}
}

func TestRemoteIndex_Fetch_NonOK_Fails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	r := NewRemoteIndex(srv.URL)
	_, err := r.Fetch()
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestRemoteIndex_Fetch_InvalidJSON_Fails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	r := NewRemoteIndex(srv.URL)
	_, err := r.Fetch()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSearchMerged_RemoteError_FailOpen(t *testing.T) {
	dir := t.TempDir()
	local := NewMarketplace(dir)
	_ = local.Add(PluginMeta{Name: "local-plugin", Backend: "local"})
	remote := NewRemoteIndex("http://127.0.0.1:0/nonexistent")
	results, err := SearchMerged(local, remote, "local")
	if err != nil {
		t.Fatalf("SearchMerged should fail-open: %v", err)
	}
	if len(results) != 1 || results[0].Name != "local-plugin" {
		t.Fatalf("expected local-plugin, got %v", results)
	}
}

func TestSearchMerged_MergesLocalAndRemote(t *testing.T) {
	dir := t.TempDir()
	local := NewMarketplace(dir)
	_ = local.Add(PluginMeta{Name: "local-gemini", Backend: "gemini"})
	body := `[{"name":"remote-qwen","backend":"qwen","description":"Qwen driver"}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer srv.Close()
	remote := NewRemoteIndex(srv.URL)
	results, err := SearchMerged(local, remote, "")
	if err != nil {
		t.Fatalf("SearchMerged: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 merged results, got %d: %v", len(results), results)
	}
}

func TestSearchMerged_DeduplicatesByName(t *testing.T) {
	dir := t.TempDir()
	local := NewMarketplace(dir)
	_ = local.Add(PluginMeta{Name: "shared", Backend: "local", Version: "1.0"})
	body := `[{"name":"shared","backend":"remote","version":"2.0"}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer srv.Close()
	remote := NewRemoteIndex(srv.URL)
	results, err := SearchMerged(local, remote, "")
	if err != nil {
		t.Fatalf("SearchMerged: %v", err)
	}
	count := 0
	for _, p := range results {
		if p.Name == "shared" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 deduped shared, got %d", count)
	}
}