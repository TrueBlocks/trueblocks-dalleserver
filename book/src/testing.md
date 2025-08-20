# Testing & Benchmarks

## Tests
Key tests in this repository:

- `request_test.go`
  - `TestParseRequest` validates request parsing, error conditions (missing series, invalid address, etc.).
  - `TestListSeries` ensures at least one series is discovered.
- `failure_test.go`
  - Simulates a generation failure by injecting a failing `generateAnnotatedImage` function and ensures handler returns 200 with the standard status message.

Run tests (skip real image generation):
```bash
make test
```
Internally sets `DALLESERVER_SKIP_IMAGE=1`.

## Race Detector
```bash
make race
```

## Benchmarks
Benchmarks (prompt/image generation path) can be run with:
```bash
make bench
```
or a focused target:
```bash
make benchmark
```

### Baseline Capture
Create timestamped JSON benchmark artifacts:
```bash
make bench-baseline
```
Outputs to `benchmarks/<timestamp>.json` and updates `benchmarks/latest.json`.
