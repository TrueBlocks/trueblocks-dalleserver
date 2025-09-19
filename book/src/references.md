# References

Primary server source files:

| Concern | File(s) |
|---------|---------|
| Entrypoint & graceful shutdown | `main.go` |
| Config (flags/env/.env) | `config.go` |
| Request parsing & series validation | `app.go` |
| Image generation orchestration | `handle_dalle.go` |
| Series listing | `handle_series.go` |
| Preview gallery | `handle_preview.go` |
| Health checks | `health.go`, `handle_health.go` |
| Metrics collection & exposition | `metrics.go`, `handle_metrics.go` |
| Middleware (logging, metrics, circuit breaker) | `middleware.go` |
| Resilience primitives | `circuit_breaker.go`, `retry.go`, `openai_client.go` |
| Errors & response contract | `errors.go` |
| Robust FS utilities | `file_operations.go` |
| Status printer (diagnostics) | `status_printer.go` |

Library documentation (prompts, phases, `DalleDress` schema):

<https://github.com/TrueBlocks/trueblocks-dalle/tree/develop/book>

Artifacts directory layout (created under the library’s data dir, typically `$HOME/.local/share/trueblocks/dalle/output`):

```
output/<series>/annotated/<address>.png
output/<series>/prompt/... (and related prompt text subfolders)
```

Future enhancements & design rationale notes live inline as comments within the corresponding Go files (search for `TODO:` or `Future` markers when exploring the codebase).
