# Testing & Benchmarks

The test suite exercises request parsing, error shaping, locking behavior, progress handling, and failure resilience without requiring real OpenAI calls.

## Modes
Image generation is skipped automatically when `OPENAI_API_KEY` is absent (or `TB_DALLE_SKIP_IMAGE=1`), enabling fast deterministic tests. The library still produces progress objects with simulated phases.

## Key Tests (Representative)
| File | Focus |
|------|-------|
| `request_test.go` | Path parsing, validation errors, query flags (`generate`, `remove`). |
| `failure_test.go` | Injected failure through `generateAnnotatedImage` stub ensures graceful 200 + error recording. |
| `health*.go` tests (if present) | Health component aggregation (filesystem, circuit breaker). |
| `metrics` related tests | Ensure counters increment on synthetic errors / retries. |

## Running
```bash
make test
```

Race detector:
```bash
make race
```

## Benchmarks
Benchmark targets measure prompt + orchestration overhead (image path still skipped unless a key is supplied). Use:
```bash
make bench
```
Focused benchmark (see `makefile`):
```bash
make benchmark
```

Baselines capture JSON artifacts for regression tracking:
```bash
make bench-baseline
```
Artifacts stored under `benchmarks/` with timestamped filenames; latest symlink / file pointer updated for easy diffing.

## Adding New Tests
* Favor table-driven tests for handlers (inputs: path + query, expected HTTP status + JSON code/message fields).
* When stubbing generation, overwrite `generateAnnotatedImage` with a closure returning a temp file path or error.
* For metrics assertions, snapshot counters before and after the action; assert monotonic increases.

## Flakiness Guidance
Avoid time-based sleeps for progress; directly call progress retrieval functions from the library after triggering a generation in tests. If unavoidable, keep sleeps short (<50ms) and document the rationale.
