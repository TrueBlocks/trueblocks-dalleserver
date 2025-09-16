# Running the Server

This section expands on invocation, flags, and shutdown behavior.

## Invocation Forms

Simplest (build + run):

```bash
make run
```

Manual (explicit build then run):

```bash
go build -o dalleserver .
./dalleserver --port=9090
```

## Flags (parsed once)

| Flag | Default | Purpose |
|------|---------|---------|
| `--port` | `8080` | Listen port (prefixed with `:` when bound). Ignored if `TB_DALLE_PORT` env var is set. |
| `--lock-ttl` | `5m` | Maximum time a (series,address) generation lock may persist (prevents stale lock starvation). |
| `--data-dir` | empty | Reserved hook for future explicit data directory configuration (delegated to library storage package). |

Note: repeated flag parsing during tests is ignored without failing.

## Environment Variables (server owned)

| Variable | Effect |
|----------|--------|
| `OPENAI_API_KEY` | Presence enables real enhancement + image fetch; absence forces `SkipImage` (mock) mode. |
| `TB_DALLE_PORT` | Overrides `--port`. |
| `TB_DALLE_SKIP_IMAGE` | Forces skip image mode even if key present. |

Environment variables consumed only by the library (e.g. enhancement timeouts, quality) are intentionally not duplicated here—see the library book.

## Skip / Mock Behavior

If no API key is detected the server still starts (emitting a warning). Progress phases execute up to the point of image acquisition which is simulated, producing annotated outputs fast for development.

## Graceful Shutdown

SIGINT / SIGTERM triggers a 10s graceful shutdown window via `http.Server.Shutdown`. In-flight requests get that window to complete; after timeout the server force closes.

## Timeouts & Protection

| Timeout | Value | Location |
|---------|-------|----------|
| ReadHeaderTimeout | 10s | Defense vs Slowloris; `http.Server` configuration. |
| ReadTimeout | 30s | Full request read bound. |
| WriteTimeout | 60s | Response write bound. |
| IdleTimeout | 120s | Keep-alive idle connections. |

OpenAI enhancement gets its own context deadline (60s) inside `openai_client.go` with an additional client-level timeout buffer.

## Status Printer

A background goroutine prints a concise table of active generations every 2s (to stderr). It is purely diagnostic and has no API surface.

## Next

See [Usage & Endpoints](./usage-endpoints.md) for request patterns, progress polling and the preview gallery.
