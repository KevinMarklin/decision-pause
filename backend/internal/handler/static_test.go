package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newStaticTestServer — сервер со статикой из временного каталога.
func newStaticTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"index.html":    "<!doctype html><title>Anti-Impulse</title><div id=root></div>",
		"assets/app.js": "console.log('ok')",
	}
	for name, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	srv := httptest.NewServer(New(&fakeService{}, Options{FrontendDist: dir}).Routes())
	t.Cleanup(srv.Close)
	return srv
}

func TestStaticServesIndexAndAssets(t *testing.T) {
	srv := newStaticTestServer(t)

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}

	resp2, err := http.Get(srv.URL + "/assets/app.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("GET /assets/app.js = %d, want 200", resp2.StatusCode)
	}
	if ct := resp2.Header.Get("Content-Type"); !strings.Contains(ct, "javascript") {
		t.Fatalf("asset Content-Type = %q, want javascript", ct)
	}
}

func TestStaticSPAFallback(t *testing.T) {
	srv := newStaticTestServer(t)

	resp, err := http.Get(srv.URL + "/history")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /history = %d, want 200 (SPA fallback)", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
}

func TestStaticMissingAssetIs404(t *testing.T) {
	srv := newStaticTestServer(t)

	resp, err := http.Get(srv.URL + "/assets/gone.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET missing asset = %d, want 404 (не index.html)", resp.StatusCode)
	}
}

func TestStaticDisabledWithoutDist(t *testing.T) {
	srv := httptest.NewServer(New(&fakeService{}, Options{}).Routes())
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET / = %d, want 404 (статика выключена)", resp.StatusCode)
	}
}
