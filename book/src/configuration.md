# Configuration

Server configuration merges (in precedence order): command‑line flags → environment variables (including those injected from `.env`) → internal defaults.

This chapter documents only variables consumed directly by server code. Library‑specific knobs (prompt enhancement timeouts, image quality, etc.) are described in the `trueblocks-dalle` book.

## .env Loader
`loadDotEnv()` reads a local `.env` file early in startup. Format: `KEY=VALUE`, comments start with `#`. Existing environment keys are never overridden. Quotes around values are stripped.

## Flags
| Flag | Default | Purpose |
|------|---------|---------|
| `--port` | `8080` | Listen port (prefixed with `:` when bound). Overridden by `TB_DALLE_PORT` if set. |
| `--lock-ttl` | `5m` | TTL for generation lock (prevents stale lock if process crashes mid-run). |
| `--data-dir` | (empty) | Reserved future hook to inject a base data directory into the library storage layer. Currently not actively used in code. |

Flags are parsed once (subsequent parsing attempts in tests are ignored silently).

## Environment Variables (Server)
| Variable | Effect |
|----------|--------|
| `OPENAI_API_KEY` | Enables real enhancement + image generation. Absence automatically sets `SkipImage=true` (mock mode). |
| `TB_DALLE_PORT` | Overrides `--port`. Value should be numeric (e.g. `9090`). |
| `TB_DALLE_SKIP_IMAGE` | Forces skip image mode even if an API key is present. Useful in tests / offline dev. |

## Derived / Implicit Behavior
| Behavior | Trigger |
|----------|---------|
| Skip image generation | `OPENAI_API_KEY` missing OR `TB_DALLE_SKIP_IMAGE=1` |
| Lock TTL fallback | Invalid `--lock-ttl` duration string → defaults to `5m` |

## Sample .env
```dotenv
# Minimal development (mock) run – leave key blank for fast iteration
# OPENAI_API_KEY=sk-...
TB_DALLE_SKIP_IMAGE=1
TB_DALLE_PORT=8080
```

## Observability Defaults
No explicit logging mode flags exist; logs are plain text via the shared `logger` package and mirrored to stderr. The metrics collector keeps only the last 1000 response time samples for percentile calculation.

## Extending Configuration
If adding new server-level toggles prefer:
1. Environment variable (document here)
2. Flag (if runtime override is operationally important)
3. Conservative default preserving existing behavior

Then expose through `Config` struct if the value needs to be widely accessed.
