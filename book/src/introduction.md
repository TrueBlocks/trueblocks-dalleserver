# Introduction

`trueblocks-dalleserver` is a thin, resilient HTTP façade around the separate
`trueblocks-dalle` library. It turns a request of the form:

```
/dalle/<series>/<ethereum-address>?generate=1
```

into an annotated PNG image plus a bundle of prompt artifacts on disk, while
surfacing live progress, structured errors, health, and lightweight metrics.

What this book covers (server only):

* Process & architecture (routing, middleware, concurrency model)
* Request parsing, validation and error response contract
* Progress polling contract returned by `/dalle/...`
* Circuit breaker + retry layers protecting OpenAI enhancement calls
* Metrics & health endpoints (JSON + Prometheus exposition)
* File system layout and robustness wrappers
* Configuration (flags, environment, `.env` loader)
* Preview gallery implementation
* Testing approach & extensibility hooks

What this book deliberately does **not** re‑document:

* Prompt assembly, attribute derivation, enhancement semantics
* The `DalleDress` or progress phase internal field-by-field description
* Series definition format or authoring guidance

For those topics, consult the upstream `trueblocks-dalle` book:

<https://github.com/TrueBlocks/trueblocks-dalle/tree/develop/book>

If reading inside a cloned mono workspace, the library book lives under the
`dalle/book` directory.

## High‑Level Flow

1. Client hits `/dalle/<series>/<addr>?generate=1`.
2. Handler validates series (against `dalle.ListSeries()`) and address format.
3. If an annotated image already exists and generation not forced, it is served directly.
4. Otherwise a background goroutine attempts `dalle.GenerateAnnotatedImage(...)` guarded by an in‑library per key lock (TTL from config) so duplicate concurrent requests coalesce.
5. While generation proceeds, repeated polls return a JSON progress document (sourced from the library `progress` package) until `done=true`.
6. On completion the PNG becomes available at the same URL (without `?generate`) and under `/files/<series>/annotated/<addr>.png` and appears in the `/preview` gallery.

All *prompt* and *image* heavy lifting is delegated to the library; this server focuses on orchestration, resilience, and presentation.

## Design Goals

| Goal | Mechanism |
|------|-----------|
| Fast no-op on cache hit | Existence check of annotated file before spawning work |
| Avoid duplicate work | Library lock keyed by (series,address) with TTL configured via `--lock-ttl` |
| Transparent progress | Library `progress.GetProgress()` snapshots serialized verbatim (plus request ID) |
| Resilience vs OpenAI hiccups | Circuit breaker + exponential backoff retry wrapper around enhancement requests |
| Operational visibility | `/metrics` (Prometheus text) and `/health` (multi-component JSON) |
| Simple ops deployment | Single binary + optional `.env` + data directory auto-create |
| Low write-time risk | Atomic file writes via temp + rename in `RobustFileOperations` |

Proceed to [Getting Started](./getting-started.md) to run the server locally.
