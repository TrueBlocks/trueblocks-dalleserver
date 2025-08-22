package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func BenchmarkGenerateAnnotatedImage(b *testing.B) {
	tmp := withTempDataDir(b, map[string]string{"bench": `{"suffix":"bench"}`})
	// Pre-create output dir to avoid timing noise
	_ = os.MkdirAll(filepath.Join(tmp, "output"), 0o750)
	dalle.ConfigureManager(dalle.ManagerOptions{MaxContexts: 5, ContextTTL: time.Minute})
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		series := "bench" // reuse same to exercise cache
		if _, err := dalle.GenerateAnnotatedImage(series, addr, true, time.Second); err != nil {
			b.Fatal(err)
		}
	}
}
