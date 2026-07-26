package main

import (
	"flag"
	"os"
	"sync"
	"time"

	"github.com/TrueBlocks/trueblocks-art/packages/creds"
)

// Config holds runtime configuration.
type Config struct {
	Port      string
	SkipImage bool
	LockTTL   time.Duration
}

var loadConfigOnce sync.Once
var cachedConfig Config

// MustLoadConfig collects configuration from flags and environment.
func MustLoadConfig() Config {
	loadConfigOnce.Do(func() {
		var cfg Config
		var portFlag string
		var lockTTLStr string
		var dataDirFlag string
		flag.StringVar(&portFlag, "port", "8080", "Port to listen on")
		flag.StringVar(&lockTTLStr, "lock-ttl", "5m", "TTL for request generation lock")
		flag.StringVar(&dataDirFlag, "data-dir", "", "Base data directory")
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
		if !creds.Has("OPENAI_API_KEY") {
			cfg.SkipImage = true
		}
		cfg.LockTTL = ttl

		// Set base data directory inside storage lazily via provided flag (environment fallback inside package).
		// storage.ConfigureDataDir(dataDirFlag)

		cachedConfig = cfg
	})
	return cachedConfig
}
