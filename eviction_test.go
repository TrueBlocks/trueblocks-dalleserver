package main

import (
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func TestContextEvictionTTL(t *testing.T) {
	// Tiny TTL and small max
	dalle.ConfigureManager(dalle.ManagerOptions{MaxContexts: 2, ContextTTL: 200 * time.Millisecond})
	seriesA := "simple"
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	if _, err := dalle.GenerateAnnotatedImage(seriesA, addr, "output", true, time.Second); err != nil {
		t.Fatal(err)
	}
	if dalle.ContextCount() != 1 {
		t.Fatalf("expected 1 context, got %d", dalle.ContextCount())
	}
	// Wait for TTL expiration
	time.Sleep(250 * time.Millisecond)
	// Trigger prune by adding another
	if _, err := dalle.GenerateAnnotatedImage("simple2", addr, "output", true, time.Second); err != nil {
		t.Fatal(err)
	}
	if dalle.ContextCount() > 2 {
		t.Fatalf("expected <=2 contexts, got %d", dalle.ContextCount())
	}
}
