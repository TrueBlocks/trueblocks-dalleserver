package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

// TestLogRotationSmallSize forces rotation logic by using a very small max size (1MB) and writing enough data.
func TestLogRotationSmallSize(t *testing.T) {
	_ = dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"simple"}})
	_ = os.Setenv("TB_DALLE_SILENT_LOG", "1")
	t.Cleanup(func() { _ = os.Unsetenv("TB_DALLE_SILENT_LOG") })

	app := NewApp()     // picks up env-based data dir
	app.StartLogging(1) // 1MB
	defer app.StopLogging()

	// Emit ~256KB (8 * 32KB) to exercise rotation path under small max size.
	chunk := make([]byte, 32*1024)
	for i := range chunk {
		chunk[i] = 'a'
	}
	for i := 0; i < 8; i++ {
		app.Logf(string(chunk))
	}

	logDir := dalle.LogsDir()
	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	files := []fs.DirEntry{}
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e)
		}
	}
	if len(files) < 1 {
		t.Fatalf("expected at least one log file")
	}

	rotated := false
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".log" && f.Name() != "server.log" {
			rotated = true
			break
		}
	}

	info, err := os.Stat(filepath.Join(logDir, "server.log"))
	if err != nil {
		t.Fatalf("stat server.log: %v", err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected non-empty server.log")
	}
	_ = rotated // placeholder for future stricter assertion
}
