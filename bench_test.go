package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

func BenchmarkGenerateAnnotatedImage(b *testing.B) {
	tmp, err := os.MkdirTemp("", "dalleserver-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = os.RemoveAll(tmp) })
	_ = os.Setenv("DALLESERVER_DATA_DIR", tmp)
	// Pre-create output dirs to avoid timing noise
	_ = os.MkdirAll(filepath.Join(tmp, "output"), 0o750)
	seriesDir := filepath.Join(tmp, "series")
	_ = os.MkdirAll(seriesDir, 0o750)
	_ = os.WriteFile(filepath.Join(seriesDir, "bench.json"), []byte(`{"suffix":"bench"}`), 0o600)
	dalle.ConfigureManager(dalle.ManagerOptions{MaxContexts: 5, ContextTTL: time.Minute})
	addr := "0xf503017d7baf7fbc0fff7492b751025c6a78179b"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		series := "bench" // reuse same to exercise cache
		if _, err := dalle.GenerateAnnotatedImage(series, addr, filepath.Join(tmp, "output"), true, time.Second); err != nil {
			b.Fatal(err)
		}
	}
}
