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
	tmp, err := os.MkdirTemp("", "dalleserver-logtest-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })
	app := &App{Config: Config{DataDir: tmp}}
	_ = os.MkdirAll(app.OutputDir(), 0o750)
	_ = os.MkdirAll(app.SeriesDir(), 0o750)
	app.ValidSeries = dalle.ListSeries(app.SeriesDir())
	app.StartLogging()
	app.Logf("test line one")
	app.Logf("test line two")
	lf := filepath.Join(tmp, "logs", "server.log")
	// lumberjack creates the file lazily on first write; give it a bit more time on slow FS
	var st os.FileInfo
	var serr error
	for i := 0; i < 50; i++ { // retry up to ~1s
		st, serr = os.Stat(lf)
		if serr == nil && st.Size() > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	// Now stop logging (close underlying rotator) after we've confirmed file creation
	app.StopLogging()
	if serr != nil || st.Size() == 0 {
		t.Fatalf("expected non-empty log file at %s (err=%v size=%d)", lf, serr, func() int64 {
			if st != nil {
				return st.Size()
			}
			return -1
		}())
	}
}
