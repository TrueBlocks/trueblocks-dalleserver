# Usage & Endpoints

All JSON responses include `success` and a correlation `request_id` (8-char UUID prefix).

## Root (`/`)
Plain text enumeration of primary endpoints. Not intended for automation.

## Series Listing (`/series`)

```
GET /series
```

Example:

```json
{
	"success": true,
	"data": {"series":["simple"], "count":1},
	## Image Generation & Progress (`/dalle/<series>/<address>`)

	Forms:

	```
	GET /dalle/<series>/<address>
	GET /dalle/<series>/<address>?generate=1
	GET /dalle/<series>/<address>?remove=1
	```

	| Condition | Plain GET | `?generate=1` | `?remove=1` |
	|-----------|-----------|---------------|-------------|
	| Annotated PNG exists | Streams PNG | Returns progress (cacheHit true) | Deletes PNG (text confirmation) |
	| PNG missing; idle | `{}` (no spawn) | Spawns generation goroutine; returns early progress or `{}` | (No-op) |
	| PNG missing; active generation | Progress JSON | Same progress (lock prevents duplicate) | (No-op) |

	Validation errors → 400 with codes: `INVALID_SERIES`, `INVALID_ADDRESS`, `MISSING_PARAMETER`.

	The progress JSON (poll until `done=true`) is produced by the library; server only adds `request_id`.

	### Locking & Concurrency
	Per-key (series,address) lock with TTL (`--lock-ttl`) coalesces concurrent generation requests. Duplicate triggers only observe progress.

	### Removal
	Only the annotated PNG is removed; prompts persist.

	## Preview Gallery (`/preview`)
	HTML template enumerating `<data>/output/<series>/annotated/*.png` grouped by series, newest first, client-side filter input.

	## Static Files (`/files/`)
	Serve any file under the output directory (read-only). Example: `/files/simple/annotated/0xabc...png`.

	## Health (`/health`)

	Modes via `check` query parameter:

	| Mode | URL | Meaning | Codes |
	|------|-----|---------|-------|
	| Full | `/health` | Composite status + components (filesystem, openai, memory, disk_space). | 200 / 503 |
	| Liveness | `/health?check=liveness` | Process responsive. | 200 |
	| Readiness | `/health?check=readiness` | Ready unless overall unhealthy. | 200 / 503 |

	OpenAI status reflects circuit breaker state (CLOSED→healthy, HALF_OPEN→degraded, OPEN→unhealthy).

	## Metrics (`/metrics`)

	| Request | Format | Purpose |
	|---------|--------|---------|
	| `/metrics` | Prometheus text | Counters, circuit breaker state gauges, response time quantiles, error breakdowns. |
	| `/metrics?format=json` | JSON | Structured snapshot (same underlying data). |

	Sample (abridged):

	```
	# Request ID: deadbeef
GET /metrics
```
Returns a placeholder Prometheus-style line:
```
	```

	## Error Shape

	```json
	{
		"success": false,
		"error": {"code":"INVALID_SERIES","message":"Invalid series name","details":"Series 'foo' not found","timestamp":1730000000,"request_id":"deadbeef"},
		"request_id": "deadbeef"
	}
	```

	Generation failures surface inside progress JSON (`error` field) with HTTP 200 to maintain polling flow.

	## Auth & CORS
	Not implemented. Apply upstream (reverse proxy / gateway) if needed.

	## Versioning
	No version prefix; additive changes preferred. Breaking changes should use new endpoints.
dalleserver_up 1
```
