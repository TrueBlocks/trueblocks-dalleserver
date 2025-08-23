package main

import (
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func BenchmarkGenerateAnnotatedImage(b *testing.B) {
	dalle.SetupTest(b, dalle.SetupTestOptions{Series: []string{"bench"}})
	dalle.ConfigureManager(dalle.ManagerOptions{MaxContexts: 5, ContextTTL: time.Minute})
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := dalle.GenerateAnnotatedImage("bench", addr, true, time.Second); err != nil {
			b.Fatal(err)
		}
	}
}
