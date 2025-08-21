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
)

// Config holds runtime configuration.
type Config struct {
	Port      string
	SkipImage bool
	LockTTL   time.Duration
	DataDir   string
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
		flag.StringVar(&dataDirFlag, "data-dir", "", "Base data directory (overrides DALLESERVER_DATA_DIR)")
		// Ignore errors (e.g., repeated parses in tests)
		if !flag.Parsed() {
			_ = flag.CommandLine.Parse(os.Args[1:])
		}
		ttl, err := time.ParseDuration(lockTTLStr)
		if err != nil {
			ttl = 5 * time.Minute
		}
		cfg.Port = ":" + portFlag
		if envPort := os.Getenv("DALLESERVER_PORT"); envPort != "" {
			cfg.Port = ":" + envPort
		}
		cfg.SkipImage = os.Getenv("DALLESERVER_SKIP_IMAGE") == "1"
		// Auto-enable skip (mock) if no API key present
		if os.Getenv("OPENAI_API_KEY") == "" {
			cfg.SkipImage = true
		}
		cfg.LockTTL = ttl
		dataDir := computeDataDir(dataDirFlag, os.Getenv("DALLESERVER_DATA_DIR"))
		if err := ensureWritable(dataDir); err != nil {
			fmt.Fprintln(os.Stderr, "FATAL: cannot use data dir:", err)
			os.Exit(1)
		}
		cfg.DataDir = dataDir
		cachedConfig = cfg
	})
	return cachedConfig
}

// computeDataDir resolves a base data directory using precedence: explicit flag > env > default (under home).
// Exposed for tests.
func computeDataDir(flagVal, envVal string) string {
	dataDir := flagVal
	if dataDir == "" {
		dataDir = envVal
	}
	if dataDir == "" {
		home, herr := os.UserHomeDir()
		if herr != nil || home == "" {
			home = "."
		}
		dataDir = filepath.Join(home, ".local", "share", "trueblocks", "dalle")
	}
	dataDir = filepath.Clean(dataDir)
	if !filepath.IsAbs(dataDir) {
		if abs, aerr := filepath.Abs(dataDir); aerr == nil {
			dataDir = abs
		}
	}
	return dataDir
}

// ensureWritable makes sure directory exists and is writable.
func ensureWritable(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	sentinel := filepath.Join(path, ".write_test")
	if werr := os.WriteFile(sentinel, []byte("ok"), 0o644); werr != nil {
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
