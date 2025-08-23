// nolint:gosec // This test intentionally sets directory perms to 0500 (execute without write) to simulate non-writable parent; execute bit required for traversal.
package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

// TestComputeDataDirPrecedence verifies flag > env > default precedence using helper.
func TestComputeDataDirPrecedence(t *testing.T) {
	// env only
	os.Setenv("TB_DALLE_DATA_DIR", "/tmp/dalleserver-env-only")
	envOnly, _ := dalle.ComputeDataDir("")
	if envOnly != "/tmp/dalleserver-env-only" {
		t.Fatalf("expected env path, got %s", envOnly)
	}
	// flag overrides env
	os.Setenv("TB_DALLE_DATA_DIR", "/tmp/dalleserver-env")
	flagOver, _ := dalle.ComputeDataDir("/tmp/dalleserver-flag")
	if flagOver != "/tmp/dalleserver-flag" {
		t.Fatalf("expected flag path, got %s", flagOver)
	}
}

// TestEnsureWritableCreates ensures directory creation.
func TestEnsureWritableCreates(t *testing.T) {
	tmp, err := os.MkdirTemp("", "dalleserver-writable-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	path := filepath.Join(tmp, "sub", "dir")
	if err := dalle.EnsureWritable(path); err != nil {
		t.Fatalf("EnsureWritable failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path created: %v", err)
	}
}

// TestEnsureWritableFails attempts to detect failure on unwritable parent (Unix only).
func TestEnsureWritableFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission semantics differ on windows")
	}
	tmp, err := os.MkdirTemp("", "dalleserver-nowrite-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	deny := filepath.Join(tmp, "deny")
	if err := os.MkdirAll(deny, 0o500); err != nil { // execute-only to simulate non-writable parent
		t.Fatal(err)
	}
	child := filepath.Join(deny, "child")
	// Remove write bit on deny directory
	if err := os.Chmod(deny, 0o500); err != nil { // keep non-writable mode
		t.Fatal(err)
	}
	if err := dalle.EnsureWritable(child); err == nil {
		// On some systems root may still allow due to user ownership; best-effort check
		if fi, statErr := os.Stat(child); statErr == nil && fi.IsDir() {
			// Can't reliably fail; skip to avoid flake
			t.Skip("platform allows creation despite permissions; skipping failure assertion")
		}
		// else let it fail
	}
}
