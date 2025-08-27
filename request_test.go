package main

import (
	"net/http"
	"net/url"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func TestParseRequest(t *testing.T) {
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"empty"}})
	app := NewApp()
	cases := []struct {
		path      string
		expectErr bool
	}{
		{"/dalle/empty/0xf503017d7baf7fbc0fff7492b751025c6a78179b", false},
		{"/dalle//0xf503017d7baf7fbc0fff7492b751025c6a78179b", true},
		{"/dalle/empty/0xdeadbeef", true},
		{"/dalle/empty/", true},
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
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"empty"}})
	list := dalle.ListSeries()
	if len(list) == 0 {
		t.Fatalf("expected at least one series")
	}
}
