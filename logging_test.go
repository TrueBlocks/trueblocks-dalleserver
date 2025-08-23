package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

// TestLoggingRotationBasic ensures the rotating logger writes the initial file.
func TestLoggingRotationBasic(t *testing.T) {
	_ = dalle.SetupTest(t, dalle.SetupTestOptions{Series: []string{"simple"}})
	_ = os.Setenv("TB_DALLE_SILENT_LOG", "1")
	t.Cleanup(func() { _ = os.Unsetenv("TB_DALLE_SILENT_LOG") })

	app := NewApp()
	app.StartLogging()
	defer app.StopLogging()
	app.Logf("test line one")
	app.Logf("test line two")
	lf := filepath.Join(dalle.LogsDir(), "server.log")
	// lumberjack creates the file lazily on first write; allow retries up to ~1s
	var st os.FileInfo
	var serr error
	for i := 0; i < 50; i++ {
		st, serr = os.Stat(lf)
		if serr == nil && st.Size() > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if serr != nil || st == nil || st.Size() == 0 {
		t.Fatalf("expected non-empty log file at %s (err=%v size=%d)", lf, serr, func() int64 {
			if st != nil {
				return st.Size()
			}
			return -1
		}())
	}
}
