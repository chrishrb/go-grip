package internal

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDirectoryListingIgnoresCacheValidators(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Hello\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-Modified-Since", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
		t.Fatalf("expected Cache-Control to disable storage, got %q", got)
	}
	if !strings.Contains(recorder.Body.String(), "README.md") {
		t.Fatalf("expected directory listing body to mention README.md, got %q", recorder.Body.String())
	}
}

func TestRegularFileStillSupportsConditionalRequests(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "plain.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write plain.txt: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/plain.txt", nil)
	req.Header.Set("If-Modified-Since", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotModified {
		t.Fatalf("expected status %d, got %d", http.StatusNotModified, recorder.Code)
	}
}

func TestMarkdownResponsesDisableCaching(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Hello\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	server := NewServer("localhost", 6419, false, false, false, NewParser())
	handler := server.newHandler(http.Dir(tmpDir))

	req := httptest.NewRequest(http.MethodGet, "/README.md", nil)
	req.Header.Set("If-Modified-Since", time.Now().Add(24*time.Hour).UTC().Format(http.TimeFormat))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); !strings.Contains(got, "no-store") {
		t.Fatalf("expected Cache-Control to disable storage, got %q", got)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html" {
		t.Fatalf("expected text/html response, got %q", got)
	}
	if !strings.Contains(recorder.Body.String(), "Hello") {
		t.Fatalf("expected rendered markdown response to contain document content, got %q", recorder.Body.String())
	}
}
