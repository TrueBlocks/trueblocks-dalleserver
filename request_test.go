package main

import (
	"net/http"
	"net/url"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v6"
)

func TestParseRequest(t *testing.T) {
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"empty"}})
	app := NewApp()
	cases := []struct {
		path         string
		expectErr    bool
		wantGenerate bool
	}{
		{"/dalle/empty/0xf503017d7baf7fbc0fff7492b751025c6a78179b", false, false},
		{"/dalle/empty/0xf503017d7baf7fbc0fff7492b751025c6a78179b?generate=1", false, true},
		{"/dalle/empty/0xf503017d7baf7fbc0fff7492b751025c6a78179b?generate=0", false, false},
		{"/dalle/empty/0xf503017d7baf7fbc0fff7492b751025c6a78179b?generate=", false, false},
		{"/dalle//0xf503017d7baf7fbc0fff7492b751025c6a78179b", true, false},
		{"/dalle/empty/0xdeadbeef", true, false},
		{"/dalle/empty/", true, false},
	}
	for _, c := range cases {
		reqUrl, _ := url.Parse(c.path)
		r := &http.Request{Method: http.MethodGet, URL: reqUrl}
		req, err := app.parseRequest(r)
		if c.expectErr && err == nil {
			t.Fatalf("expected error for %s", c.path)
		}
		if !c.expectErr && err != nil {
			t.Fatalf("unexpected error for %s: %v", c.path, err)
		}
		if !c.expectErr && req.generate != c.wantGenerate {
			t.Fatalf("generate mismatch for %s: got %v want %v", c.path, req.generate, c.wantGenerate)
		}
	}
}

func TestListSeries(t *testing.T) {
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"empty"}})
	list := dalle.ListSeries()
	if len(list) == 0 {
		t.Fatalf("expected at least one series")
	}
}
