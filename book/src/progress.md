# Progress & Polling Contract

The server exposes *progress* transparently by forwarding the library’s snapshot for an in‑flight generation. This chapter focuses on how clients should consume it rather than restating the library schema.

## Lifecycle Phases (High Level)
Typical sequence (library defined):

```
setup → base_prompts → enhance_prompt → image_prep → image_wait → image_download → annotate → completed
```

Phases may be marked `skipped` (e.g. enhancement when running in skip mode).

## Polling Pattern
1. Kick off (or re-use) generation:
   ```
   GET /dalle/<series>/<address>?generate=1
   ```
2. Poll without changing parameters (optionally retain `?generate=1`; lock prevents duplicate work) once per second until `done=true`.
3. When `done=true` and no `error`, re-request **without** `generate` to stream the final PNG directly.

Fish loop example:
```fish
while true
    curl -s "http://localhost:8080/dalle/simple/<addr>?generate=1" | jq '.percent, .current, .done'
    sleep 1
end
```

## Percent & ETA
`percent` and `etaSeconds` reflect exponential moving averages (EMA) tracked across prior successful runs (cache hits excluded). Expect more accuracy as the system observes more generations.

## Cache Hits
If the annotated PNG already exists before generation starts: snapshot returns quickly with `cacheHit=true`, `done=true`, minimal or empty phase timings.

## Error Semantics
Progress snapshot may include a non-empty `error` while `done=true`; client should surface the message and avoid retry loops unless user refires manually (e.g. after clearing causes like rate limiting). HTTP status still 200 in this case—polling contract relies on payload state not transport errors.

## Stability & Forward Compatibility
The progress object is additive: new fields may appear; existing names are stable. Clients should ignore unknown fields.

## Where to Find Field Details
For a full enumeration of snapshot and `DalleDress` fields, consult the `trueblocks-dalle` book (progress section). This server layer intentionally avoids duplicating that documentation to prevent drift.

## Observability Coupling
Response time metrics collected in middleware are independent of progress timing; phase averages feed only the progress percent/ETA calculation inside the library.

## Future Extensions (Server-Level)
* Optional WebSocket push (reduces polling overhead)
* Soft cancellation endpoint (e.g. `DELETE /dalle/<series>/<address>`)
* Streaming Server-Sent Events wrapper around snapshot diffs
