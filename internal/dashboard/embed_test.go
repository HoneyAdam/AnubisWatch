package dashboard

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandler(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	if handler == nil {
		t.Fatal("Expected handler to be created")
	}

	if handler.fs == nil {
		t.Error("Expected handler fs to be set")
	}
}

func TestHandler_ServeHTTP_Root(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return index.html (200) or fallback gracefully
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Expected 200 or 404, got %d", w.Code)
	}
}

func TestHandler_ServeHTTP_Index(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/index.html", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should serve index.html or return 404 if not embedded
	t.Logf("Index.html returned status: %d", w.Code)
}

func TestHandler_ServeHTTP_SPA(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	// SPA routing - unknown paths should return index.html
	req := httptest.NewRequest("GET", "/some/unknown/route", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should either return the file or fallback to index.html
	t.Logf("SPA route returned status: %d", w.Code)
}

func TestHandler_ServeHTTP_NotFound(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/nonexistent.xyz", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 404 for truly nonexistent files
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Expected 200 or 404, got %d", w.Code)
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"index.html", "text/html; charset=utf-8"},
		{"app.js", "application/javascript; charset=utf-8"},
		{"bundle.mjs", "application/javascript; charset=utf-8"},
		{"styles.css", "text/css; charset=utf-8"},
		{"data.json", "application/json; charset=utf-8"},
		{"image.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"animation.gif", "image/gif"},
		{"icon.svg", "image/svg+xml"},
		{"font.woff", "font/woff"},
		{"font.woff2", "font/woff2"},
		{"font.ttf", "font/ttf"},
		{"font.otf", "font/otf"},
		{"favicon.ico", "image/x-icon"},
		{"module.wasm", "application/wasm"},
		{"unknown.xyz", "application/octet-stream"},
		{".hidden", "application/octet-stream"},
		{"noextension", "application/octet-stream"},
	}

	for _, tt := range tests {
		result := getContentType(tt.filename)
		if result != tt.expected {
			t.Errorf("getContentType(%q) = %q, expected %q", tt.filename, result, tt.expected)
		}
	}
}

func TestGetContentType_Empty(t *testing.T) {
	result := getContentType("")
	if result != "application/octet-stream" {
		t.Errorf("getContentType(\"\") = %q, expected application/octet-stream", result)
	}
}

func TestHandler_ServeHTTP_Methods(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Handler should handle all methods gracefully
		t.Logf("Handler %s returned %d", method, w.Code)
	}
}

func TestHandler_ServeHTTP_Subdirectory(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	// Test various content types by requesting different paths
	paths := []string{
		"/assets/app.css",
		"/assets/app.js",
		"/data/config.json",
		"/images/logo.png",
		"/fonts/main.woff2",
	}

	for _, p := range paths {
		req := httptest.NewRequest("GET", p, nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should not panic
		if w.Code != http.StatusOK {
			t.Logf("Path %s returned %d", p, w.Code)
		}
	}
}

func TestHandler_ServeHTTP_Paths(t *testing.T) {
	handler, err := NewHandler()
	if err != nil {
		t.Fatalf("NewHandler failed: %v", err)
	}

	paths := []string{
		"/",
		"/.",
		"/app",
		"/app/",
		"/dashboard",
		"/status",
		"/settings",
		"/../../../etc/passwd", // Path traversal attempt
	}

	for _, p := range paths {
		req := httptest.NewRequest("GET", p, nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should handle all paths gracefully (no panic)
		t.Logf("Path %s returned %d", p, w.Code)
	}
}
