package main

import (
	"bufio"
	"flag"
	"os"
	"strings"
	"sync"
	"time"
)

// Config holds runtime configuration.
type Config struct {
	Port      string
	SkipImage bool
	LockTTL   time.Duration
	LogJSON   bool
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
		var logJSON bool
		flag.StringVar(&portFlag, "port", "8080", "Port to listen on")
		flag.StringVar(&lockTTLStr, "lock-ttl", "5m", "TTL for request generation lock")
		flag.BoolVar(&logJSON, "log-json", true, "Emit logs in JSON format")
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
		cfg.LogJSON = logJSON
		cachedConfig = cfg
	})
	return cachedConfig
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
