package main

import (
	"sync"
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

// TestConcurrentGenerate ensures lock prevents redundant heavy work; we just assert no errors and same path.
func TestConcurrentGenerate(t *testing.T) {
	dalle.ConfigureManager(dalle.ManagerOptions{MaxContexts: 5, ContextTTL: time.Minute})
	dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"empty"}})
	series := "empty"
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make(chan error, n)
	paths := make(chan string, n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			p, err := dalle.GenerateAnnotatedImage(series, addr, true, 2*time.Second)
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
