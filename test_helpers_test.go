package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

// withTempDataDir creates an isolated temporary data directory, sets the
// TB_DALLE_DATA_DIR env var (cleaned up automatically), optionally seeds
// series definition JSON files, and returns the base path.
// seriesFiles map key may be either the bare series name or the filename
// (with or without .json). Value is the JSON content. Passing nil/empty map
// skips seeding. Works for both *testing.T and *testing.B via testing.TB.
func withTempDataDir(tb testing.TB, seriesFiles map[string]string) string {
	tb.Helper()
	tmp, err := os.MkdirTemp("", "dalleserver-test-*")
	if err != nil {
		tb.Fatal(err)
	}
	tb.Cleanup(func() { _ = os.RemoveAll(tmp) })
	_ = os.Setenv("TB_DALLE_DATA_DIR", tmp)
	if err := dalle.InitDataDir(""); err != nil {
		// Fail fast if initialization fails
		os.Unsetenv("TB_DALLE_DATA_DIR")
		tb.Fatalf("InitDataDir failed: %v", err)
	}
	tb.Cleanup(func() { os.Unsetenv("TB_DALLE_DATA_DIR") })
	if len(seriesFiles) > 0 {
		seriesDir := filepath.Join(tmp, "series")
		_ = os.MkdirAll(seriesDir, 0o750)
		for name, content := range seriesFiles {
			if !strings.HasSuffix(name, ".json") {
				name += ".json"
			}
			_ = os.WriteFile(filepath.Join(seriesDir, name), []byte(content), 0o600)
		}
	}
	return tmp
}
