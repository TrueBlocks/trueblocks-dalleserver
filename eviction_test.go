package main

import (
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func TestContextEvictionTTL(t *testing.T) {
	dalle.ConfigureManager(dalle.ManagerOptions{MaxContexts: 2, ContextTTL: 200 * time.Millisecond})
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"empty", "empty2"}})
	seriesA := "empty"
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	if _, err := dalle.GenerateAnnotatedImage(seriesA, addr, true, time.Second); err != nil {
		t.Fatal(err)
	}
	if dalle.ContextCount() != 1 {
		t.Fatalf("expected 1 context, got %d", dalle.ContextCount())
	}
	// Wait for TTL expiration
	time.Sleep(250 * time.Millisecond)
	// Trigger prune by adding another
	if _, err := dalle.GenerateAnnotatedImage("empty2", addr, true, time.Second); err != nil {
		t.Fatal(err)
	}
	if dalle.ContextCount() > 2 {
		t.Fatalf("expected <=2 contexts, got %d", dalle.ContextCount())
	}
}
