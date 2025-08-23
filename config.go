package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
)

// Config holds runtime configuration.
type Config struct {
	Port      string
	SkipImage bool
	LockTTL   time.Duration
	SoonToGo  string
}

// LoadConfig collects configuration from flags and environment.
var loadConfigOnce sync.Once
var cachedConfig Config

func LoadConfig() Config {
	loadConfigOnce.Do(func() {
		loadDotEnv()
		var cfg Config
		var portFlag string
		var lockTTLStr string
		var dataDirFlag string
		flag.StringVar(&portFlag, "port", "8080", "Port to listen on")
		flag.StringVar(&lockTTLStr, "lock-ttl", "5m", "TTL for request generation lock")
		flag.StringVar(&dataDirFlag, "data-dir", "", "Base data directory (overrides TB_DALLE_DATA_DIR)")
		// Ignore errors (e.g., repeated parses in tests)
		if !flag.Parsed() {
			_ = flag.CommandLine.Parse(os.Args[1:])
		}
		ttl, err := time.ParseDuration(lockTTLStr)
		if err != nil {
			ttl = 5 * time.Minute
		}
		cfg.Port = ":" + portFlag
		if envPort := os.Getenv("TB_DALLE_PORT"); envPort != "" {
			cfg.Port = ":" + envPort
		}
		cfg.SkipImage = os.Getenv("TB_DALLE_SKIP_IMAGE") == "1"
		// Auto-enable skip (mock) if no API key present
		if os.Getenv("OPENAI_API_KEY") == "" {
			cfg.SkipImage = true
		}
		cfg.LockTTL = ttl
		dataDir := dalle.ComputeDataDir(dataDirFlag, os.Getenv("TB_DALLE_DATA_DIR"))
		if err := ensureWritable(dataDir); err != nil {
			// Fall back to a temp directory instead of exiting so tests / server can continue.
			tmp, terr := os.MkdirTemp("", "dalleserver-fallback-*")
			if terr != nil {
				fmt.Fprintln(os.Stderr, "ERROR: cannot establish writable data dir:", err)
				// Last resort: keep original (likely failing) path to surface errors later.
				dataDir = dataDir + "-unwritable"
			} else {
				fmt.Fprintln(os.Stderr, "WARNING: using fallback temp data dir due to error:", err)
				dataDir = tmp
			}
		}
		cfg.SoonToGo = dataDir
		cachedConfig = cfg
	})
	return cachedConfig
}

// ensureWritable makes sure directory exists and is writable.
func ensureWritable(path string) error {
	// Create (or ensure) the directory with restricted permissions; callers can relax if explicitly required.
	if err := os.MkdirAll(path, 0o750); err != nil {
		return err
	}
	sentinel := filepath.Join(path, ".write_test")
	// Use 0o600 for the write test to satisfy gosec and to avoid exposing potential sensitive data.
	if werr := os.WriteFile(sentinel, []byte("ok"), 0o600); werr != nil {
		return werr
	}
	_ = os.Remove(sentinel)
	return nil
}

// loadDotEnv loads key=value pairs from a local .env file (simple parser) if present.
// Lines beginning with # are ignored. Keys already present in environment are not overwritten.
func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if eq := strings.Index(line, "="); eq > 0 {
			k := strings.TrimSpace(line[:eq])
			v := strings.TrimSpace(line[eq+1:])
			if _, exists := os.LookupEnv(k); !exists {
				_ = os.Setenv(k, strings.Trim(v, `"`))
			}
		}
	}
}
