package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test error paths on the /dalle/ handler surface proper 400s with expected messages.
func TestHandleDalleErrors(t *testing.T) {
	tmp, err := os.MkdirTemp("", "dalleserver-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })
	_ = os.Setenv("DALLESERVER_DATA_DIR", tmp)
	seriesDir := filepath.Join(tmp, "series")
	_ = os.MkdirAll(seriesDir, 0o750)
	_ = os.WriteFile(filepath.Join(seriesDir, "simple.json"), []byte(`{"suffix":"simple"}`), 0o600)
	app := NewApp()
	app.StartLogging()
	defer app.StopLogging()

	cases := []struct {
		name       string
		path       string
		wantStatus int
		wantSubstr string
	}{
		{"invalid series", "/dalle/notaseries/0xf503017d7baf7fbc0fff7492b751025c6a78179b", http.StatusBadRequest, "invalid series"},
		{"malformed address", "/dalle/simple/0xdeadbeef", http.StatusBadRequest, "invalid address"},
		{"missing address", "/dalle/simple/", http.StatusBadRequest, "address is required"},
		{"missing both", "/dalle/", http.StatusBadRequest, "invalid request"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, c.path, nil)
			w := httptest.NewRecorder()
			app.handleDalleDress(w, r)
			res := w.Result()
			if res.StatusCode != c.wantStatus {
				t.Fatalf("%s: expected status %d got %d", c.name, c.wantStatus, res.StatusCode)
			}
			bodyBytes, _ := io.ReadAll(res.Body)
			body := string(bodyBytes)
			if c.wantSubstr != "" && !strings.Contains(body, c.wantSubstr) {
				t.Fatalf("%s: expected body to contain %q; got %q", c.name, c.wantSubstr, body)
			}
		})
	}
}
