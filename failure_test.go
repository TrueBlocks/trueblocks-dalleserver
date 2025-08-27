package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

// TestSimulatedOpenAIFailure injects a failing generateAnnotatedImage to ensure the handler
// logs the error path without panicking and still responds 200 with standard message.
func TestSimulatedOpenAIFailure(t *testing.T) {
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"empty"}})
	app := NewApp()
	app.StartLogging()
	defer app.StopLogging()

	// Prepare injection
	called := 0
	original := generateAnnotatedImage
	generateAnnotatedImage = func(series, addr string, skip bool, ttl time.Duration) (string, error) {
		called++
		return "", fmt.Errorf("forced failure for testing")
	}
	defer func() { generateAnnotatedImage = original }()

	// Force synchronous path
	prevDebug := isDebugging
	isDebugging = true
	defer func() { isDebugging = prevDebug }()

	// Use a unique address unlikely to have prior progress to better isolate this test
	addr := "0x5555555555555555555555555555555555555555"
	url := "/dalle/empty/" + addr + "?generate=1"
	r := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	app.handleDalleDress(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	trimmed := strings.TrimSpace(body)
	// Handler now always returns JSON progress (or {} if no progress yet) instead of a static message.
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}\n") && !strings.HasSuffix(trimmed, "}") {
		// Basic sanity check that we received JSON; we don't assert specific fields because
		// the injected failure prevents progress initialization.
		t.Fatalf("expected JSON progress body, got %q", body)
	}
	if called != 1 {
		t.Fatalf("expected injected generator to be called once, got %d", called)
	}
}
