# Progress & Metrics

The `/dalle/<series>/<address>` endpoint now returns structured JSON snapshots during generation instead of a plain status line. Each snapshot embeds the live `DalleDress` object and a timeline of canonical phases:

```
setup → base_prompts → enhance_prompt → image_prep → image_wait → image_download → annotate → completed
```

## Snapshot Schema (example)

```json
{
  "series": "simple",
  "address": "0xabc...",
  "currentPhase": "image_wait",
  "startedNs": 1730000000000000000,
  "percent": 37.2,
  "etaSeconds": 12.4,
  "done": false,
  "error": "",
  "cacheHit": false,
  "phases": [
    {"name":"setup","startedNs":1730000000000000000,"endedNs":1730000001000000000,"skipped":false,"error":""},
    {"name":"base_prompts","startedNs":1730000001000000000,"endedNs":1730000002000000000,"skipped":false,"error":""},
    {"name":"enhance_prompt","startedNs":1730000002000000000,"endedNs":1730000003000000000,"skipped":true,"error":""},
    {"name":"image_prep","startedNs":1730000003000000000,"endedNs":0,"skipped":false,"error":""}
  ],
  "dalleDress": { /* extended fields; never omitted */ },
  "phaseAverages": { "image_wait": 2500000000 }
}
```

### Field Notes
- All top-level keys are always present (no `omitempty`). Empty slices are `[]` not `null`.
- `percent` and `etaSeconds` are computed from exponential moving averages (EMA, alpha=0.2) of prior successful, non-skipped runs.
- During a phase, elapsed time is capped at its average when computing percent to avoid overshoot.
- `cacheHit=true` denotes the annotated image already existed; such runs do not update EMAs or `generationRuns`.

## Polling Pattern
Perform an initial request with `?generate=1` (if generation is desired) and poll the same URL until `done=true`.

Example (bash/fish):
```bash
while true; sleep 1; curl -s http://localhost:8080/dalle/simple/0x....?generate=1 | jq '.percent, .currentPhase, .done' ; end
```

## Metrics Persistence
Global rolling stats persist to:
```
metrics/progress_phase_stats.json
```
Schema (version `v1`):
```json
{
  "version": "v1",
  "phaseAverages": {
    "image_wait": {"count": 4, "avgNs": 2100000000}
  },
  "generationRuns": 12,
  "cacheHits": 5
}
```
Only successful, non-skipped, non-cache-hit phase durations feed the EMA. Cache hits increment `cacheHits` only.

## Testing Helpers
Inside the `dalle` module:
- `GetProgress(series,address)` – retrieve a snapshot (deletes it once `done=true`).
- `ResetMetricsForTest()` – clear in-memory and on-disk metrics.
- `ForceMetricsSave()` – flush current metrics to disk.

## Error Handling
If a phase fails, the run transitions to `completed` with `error` populated; failed phase duration is excluded from averages.

## Concurrency & Integrity
A single `ProgressManager` serializes updates per `(series,address)`. The live `DalleDress` pointer is reused—treat it as read-only when consumed in handlers or tests.

## Cache Hit Short-Circuit
If `output/<series>/annotated/<address>.png` exists before generation begins, a synthetic completed snapshot is emitted immediately (if no active run), flagged `cacheHit=true`.

## Future Enhancements (Ideas)
- Optional run archive (disabled currently) writing each completed snapshot to disk.
- Rolling p95 / p99 latency tracking alongside EMA.
- WebSocket push for progress (avoiding polling).
- Percent smoothing at phase boundaries to avoid visual jumps.
