# Usage & Endpoints

## Root
```
/
```
Lists available endpoints:
- `/series` (and `/series/`)
- `/dalle/<series>/<address>?[generate|remove]`
- `/preview`
- `/healthz`
- `/metrics`

## List Series
```
GET /series
```
Returns: `Available series: [ ... ]` as a JSON array (pretty-printed inside the line).

Implementation: `handleSeries` calls `dalle.ListSeries("output")` and prints the list.

## Generate or Fetch Image
```
GET /dalle/<series>/<address>
GET /dalle/<series>/<address>?generate=1
GET /dalle/<series>/<address>?remove=1
```
- Without query flags and if the annotated PNG exists, the image is served.
- `?generate` triggers asynchronous generation (returns immediately with a status line).
- `?remove` deletes the existing annotated image if present.

Response when (re)generation starts:
```
Your image will be ready shortly
```

If removal succeeds:
```
Image removed
```
If removal requested but file absent:
```
Image not found
```

### Address Validation
- Series must be one returned by `/series` (validated in `parseRequest`).
- Address must be a valid Ethereum address (EIP-55 checksum normalization applied).

### Concurrency
Generation is launched in a goroutine unless `isDebugging` is set (then synchronous). Progress is logged; no streaming output is provided.

## Preview Gallery
```
GET /preview
```
Renders an HTML page grouping annotated images by series with a client-side filter box. Images are displayed square using a CSS aspect-ratio wrapper. Timestamps (mod time) are shown under each address.

## Static Files
```
GET /files/<path under output/>
```
Serves raw files from the `output/` tree (e.g. annotated images).

## Health
```
GET /healthz
```
Returns JSON: `{"status":"ok"}`

## Metrics
```
GET /metrics
```
Returns a placeholder Prometheus-style line:
```
dalleserver_up 1
```
