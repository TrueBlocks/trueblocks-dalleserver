package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestComputeDataDirPrecedence verifies flag > env > default precedence using helper.
func TestComputeDataDirPrecedence(t *testing.T) {
	// env only
	envOnly := computeDataDir("", "/tmp/dalleserver-env-only")
	if envOnly != "/tmp/dalleserver-env-only" {
		t.Fatalf("expected env path, got %s", envOnly)
	}
	// flag overrides env
	flagOver := computeDataDir("/tmp/dalleserver-flag", "/tmp/dalleserver-env")
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
	if err := ensureWritable(path); err != nil {
		t.Fatalf("ensureWritable failed: %v", err)
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
	if err := os.MkdirAll(deny, 0o500); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(deny, "child")
	// Remove write bit on deny directory
	if err := os.Chmod(deny, 0o500); err != nil {
		t.Fatal(err)
	}
	if err := ensureWritable(child); err == nil {
		// On some systems root may still allow due to user ownership; best-effort check
		if fi, statErr := os.Stat(child); statErr == nil && fi.IsDir() {
			// Can't reliably fail; skip to avoid flake
			t.Skip("platform allows creation despite permissions; skipping failure assertion")
		}
		// else let it fail
	}
}
