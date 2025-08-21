package main

import (
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func BenchmarkGenerateAnnotatedImage(b *testing.B) {
	dalle.ConfigureManager(dalle.ManagerOptions{MaxContexts: 5, ContextTTL: time.Minute})
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		series := "bench" // reuse same to exercise cache
		// OUTPUT_DIR
		if _, err := dalle.GenerateAnnotatedImage(series, addr, "output", true, time.Second); err != nil {
			b.Fatal(err)
		}
	}
}
