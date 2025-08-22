package main

import (
	"net/http"
	"net/url"
	"path/filepath"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func TestParseRequest(t *testing.T) {
	_ = withTempDataDir(t, map[string]string{"simple": `{"suffix":"simple"}`})
	// Temp data dir seeded with series; proceed.
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
	tmp2 := withTempDataDir(t, map[string]string{"simple": `{"suffix":"simple"}`})
	seriesDir := filepath.Join(tmp2, "series")
	list := dalle.ListSeries(seriesDir)
	if len(list) == 0 {
		t.Fatalf("expected at least one series")
	}
}
