package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func TestContextEvictionTTL(t *testing.T) {
	// Tiny TTL and small max
	dalle.ConfigureManager(dalle.ManagerOptions{MaxContexts: 2, ContextTTL: 200 * time.Millisecond})
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
	_ = os.WriteFile(filepath.Join(seriesDir, "simple2.json"), []byte(`{"suffix":"simple2"}`), 0o600)
	seriesA := "simple"
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	if _, err := dalle.GenerateAnnotatedImage(seriesA, addr, filepath.Join(tmp, "output"), true, time.Second); err != nil {
		t.Fatal(err)
	}
	if dalle.ContextCount() != 1 {
		t.Fatalf("expected 1 context, got %d", dalle.ContextCount())
	}
	// Wait for TTL expiration
	time.Sleep(250 * time.Millisecond)
	// Trigger prune by adding another
	if _, err := dalle.GenerateAnnotatedImage("simple2", addr, filepath.Join(tmp, "output"), true, time.Second); err != nil {
		t.Fatal(err)
	}
	if dalle.ContextCount() > 2 {
		t.Fatalf("expected <=2 contexts, got %d", dalle.ContextCount())
	}
}
