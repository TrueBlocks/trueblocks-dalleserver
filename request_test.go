package main

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func TestParseRequest(t *testing.T) {
	tmp, err := os.MkdirTemp("", "dalleserver-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })
	_ = os.Setenv("DALLESERVER_DATA_DIR", tmp)
	// Ensure basic dirs
	seriesDir := filepath.Join(tmp, "series")
	_ = os.MkdirAll(seriesDir, 0o750)
	_ = os.WriteFile(filepath.Join(seriesDir, "simple.json"), []byte(`{"suffix":"simple"}`), 0o600)
	app := NewApp()
	cases := []struct {
		path      string
		expectErr bool
	}{
		{"/dalle/simple/0xf503017d7baf7fbc0fff7492b751025c6a78179b", false},
		{"/dalle//0xf503017d7baf7fbc0fff7492b751025c6a78179b", true},
		{"/dalle/simple/0xdeadbeef", true},
		{"/dalle/simple/", true},
	}
	for _, c := range cases {
		reqUrl, _ := url.Parse(c.path)
		r := &http.Request{Method: http.MethodGet, URL: reqUrl}
		_, err := app.parseRequest(r)
		if c.expectErr && err == nil {
			t.Fatalf("expected error for %s", c.path)
		}
		if !c.expectErr && err != nil {
			t.Fatalf("unexpected error for %s: %v", c.path, err)
		}
	}
}

func TestListSeries(t *testing.T) {
	tmp, err := os.MkdirTemp("", "dalleserver-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })
	_ = os.Setenv("DALLESERVER_DATA_DIR", tmp)
	seriesDir := filepath.Join(tmp, "series")
	_ = os.MkdirAll(seriesDir, 0o750)
	_ = os.WriteFile(filepath.Join(seriesDir, "simple.json"), []byte(`{"suffix":"simple"}`), 0o600)
	list := dalle.ListSeries(seriesDir)
	if len(list) == 0 {
		t.Fatalf("expected at least one series")
	}
}
