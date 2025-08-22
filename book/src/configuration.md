# Configuration

Configuration combines command-line flags, environment variables, and a `.env` file.

## `.env` Loading
A simple loader (`loadDotEnv`) reads `.env` at startup. Lines of the form `KEY=VALUE` are added to the environment if not already set.

## Command-line Flags
| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8080` | HTTP listen port (overridden by env `TB_DALLE_PORT`) |
| `--lock-ttl` | `5m` | TTL for the in-memory generation lock |
| `--log-json` | `true` | Emit logs in JSON format (passed into config) |

## Environment Variables
| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | Required for enhancement & image generation (server exits early if unset). |
| `TB_DALLE_PORT` | Alternate way to set port (e.g. `8090`). |
| `TB_DALLE_SKIP_IMAGE` | If set to `1` forces skip of image generation (mock mode). |
| `TB_DALLE_NO_ENHANCE` | If set to `1` instructs downstream enhancement logic to skip (handled in library). |
| `TB_DALLE_ENHANCE_TIMEOUT` | Enhance step timeout (default 60s in code). |
| `TB_DALLE_IMAGE_TIMEOUT` | Image request + download timeout (default 30s). |
| `DALLE_QUALITY` | Value forwarded to image generation (e.g. `standard`). |

## Auto Skip Behavior
If `OPENAI_API_KEY` is *not* present, `SkipImage` is automatically set to true in config (mock/offline mode). The main program then still enforces presence of the key and exits before serving unless you explicitly exported a key (see `main.go`).
