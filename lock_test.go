package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

// TestConcurrentGenerate ensures lock prevents redundant heavy work; we just assert no errors and same path.
func TestConcurrentGenerate(t *testing.T) {
	// Configure small manager to ensure eviction not triggered here
	dalle.ConfigureManager(dalle.ManagerOptions{MaxContexts: 5, ContextTTL: time.Minute})
	tmp, err := os.MkdirTemp("", "dalleserver-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })
	_ = os.Setenv("DALLESERVER_DATA_DIR", tmp)
	_ = os.MkdirAll(filepath.Join(tmp, "output"), 0o750)
	seriesDir := filepath.Join(tmp, "series")
	_ = os.MkdirAll(seriesDir, 0o750)
	_ = os.WriteFile(filepath.Join(seriesDir, "simple.json"), []byte(`{"suffix":"simple"}`), 0o600)
	series := "simple"
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	paths := make(chan string, n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			p, err := dalle.GenerateAnnotatedImage(series, addr, filepath.Join(tmp, "output"), true, 2*time.Second)
			if err != nil {
				errs <- err
				return
			}
			paths <- p
		}()
	}
	wg.Wait()
	close(errs)
	close(paths)
	for e := range errs {
		if e != nil {
			t.Fatalf("unexpected error: %v", e)
		}
	}
	var first string
	for p := range paths {
		if first == "" {
			first = p
		} else if p != first {
			t.Fatalf("paths differ: %s vs %s", first, p)
		}
	}
}
