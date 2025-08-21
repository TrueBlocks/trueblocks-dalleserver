package main

// Instrumentation TestMain to diagnose unexplained non-zero exit status where all tests appear to pass.
// If the root test suite returns a non-zero code, we dump goroutine stacks and basic diagnostics.

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[instrument] TestMain start at %s (go=%s)\n", start.Format(time.RFC3339Nano), runtime.Version())
	code := m.Run()
	dur := time.Since(start)
	fmt.Fprintf(os.Stderr, "[instrument] TestMain m.Run complete code=%d duration=%s\n", code, dur)
	if code != 0 {
		fmt.Fprintf(os.Stderr, "[instrument] Non-zero exit detected: dumping goroutines...\n")
		// Goroutine dump (level 2 to include stack frames)
		if gr := pprof.Lookup("goroutine"); gr != nil {
			_ = gr.WriteTo(os.Stderr, 2)
		}
		// Memory stats
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		fmt.Fprintf(os.Stderr, "[instrument] MemStats: Alloc=%d Sys=%d NumGC=%d Goroutines=%d\n", ms.Alloc, ms.Sys, ms.NumGC, runtime.NumGoroutine())
	}
	os.Exit(code)
}
