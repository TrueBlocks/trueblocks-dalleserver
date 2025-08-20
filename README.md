# trueblocks-dalleserver

An example HTTP server demonstrating how to build image experiences on top of the
[`trueblocks-dalle`](https://github.com/TrueBlocks/trueblocks-dalle) Go package.

It turns Ethereum addresses into consistently generated, stylistically filtered ("series") images.
This repo shows a server approach; a companion desktop / direct usage example can build images
without HTTP by calling the same library functions.

![CI](https://github.com/TrueBlocks/trueblocks-dalleserver/actions/workflows/ci.yml/badge.svg?branch=develop)

## Why this repo exists

Developers often ask: "How do I actually use the `trueblocks-dalle` module in a program?"

This code answers that by demonstrating:

* Discovering and validating a "series" (prompt filter set) at runtime
* Generating prompts (data / title / terse / full / enhanced) for an address
* Optionally enhancing prompts via OpenAI (LLM) with timeouts + fallbacks
* Generating and annotating DALL·E images (or skipping in offline mode)
* Caching + locking so concurrent requests don’t stampede
* Exposing a REST API and a gallery preview page
* Testing, linting, benchmarking and baseline capture

If you only need library usage, jump to [Direct library usage](#direct-library-usage).

## Quick start

```bash
git clone https://github.com/TrueBlocks/trueblocks-dalleserver.git # or: git clone git@github.com:TrueBlocks/trueblocks-dalleserver.git
cd trueblocks-dalleserver
cp .env.example .env            # create and edit (.env is auto-loaded)
make run                        # builds then starts :8080
open http://localhost:8080/preview
```

List available series:

```bash
curl http://localhost:8080/series
```

Fetch (or trigger) an image:

```bash
open "http://localhost:8080/dalle/simple/0xf503017d7baf7fbc0fff7492b751025c6a78179b?generate=1"
```

Preview gallery:

```
http://localhost:8080/preview
```

## Make targets

| Command               | Purpose |
|-----------------------|---------|
| `make run`            | Build + run server on :8080 |
| `make lint`           | Install + run golangci-lint (pinned) |
| `make test`           | Run tests (image generation skipped) |
| `make race`           | Run race detector tests (skip image) |
| `make bench`          | Run all benchmarks (skip image) |
| `make benchmark`      | Focused benchmark target |
| `make bench-baseline` | Produce timestamped JSON benchmark artifacts (benchmarks/*.json) |
| `make clean-output`   | Remove generated PNGs |

## Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | Key for enhancement + DALL·E image calls (required unless skipping) | (none) |
| `DALLESERVER_SKIP_IMAGE` | `1` to skip actual image generation (offline / fast tests) | unset |
| `DALLESERVER_NO_ENHANCE` | `1` to disable LLM enhancement (use raw prompt) | unset |
| `DALLESERVER_ENHANCE_TIMEOUT` | Override enhance prompt timeout (e.g. `75s`) | `60s` |
| `DALLESERVER_IMAGE_TIMEOUT` | Timeout for image request + download | `30s` |
| `DALLE_QUALITY` | DALL·E quality parameter (`standard`, `hd`, etc.) | `standard` |

Example (fish shell):

```fish
set -x OPENAI_API_KEY "sk-..."
make run
```

Or use a local `.env` file (preferred for development):

```bash
cp .env.example .env
edit .env  # populate OPENAI_API_KEY and options
make run
```

Offline/dev mode:

```fish
set -x DALLESERVER_SKIP_IMAGE 1; make run
```

## Endpoints

| Path | Description |
|------|-------------|
| `/dalle/<series>/<address>` | Returns image if already generated; else a message. Add `?generate=1` to force generation. |
| `/series` | Lists available series names. |
| `/preview` | HTML gallery of annotated images (filterable). |
| `/files/...` | Static access to generated output tree. |
| `/healthz` | Basic health probe JSON. |
| `/metrics` | Placeholder metrics endpoint. |

## Output structure

```
output/
  series/                 # JSON series definitions (filters)
  <series>/
  data/                 # Raw data prompt
  title/                # Title prompt
  terse/                # Short prompt
  prompt/               # Full prompt
  enhanced/             # Enhanced (LLM) prompt text
  annotated/            # Final PNG images (watermarked)
```

Transient lock files live under `pending/` while generation is occurring.

## Direct library usage

If you want to skip the server and just integrate image generation, import the package:

```go
import (
  dalle "github.com/TrueBlocks/trueblocks-dalle/v2"
  "time"
)

func generateOne(series, addr string) error {
  // outputDir must exist / will be created relative to cwd
  _, err := dalle.GenerateAnnotatedImage(series, addr, "output", false /* skipImage */, 30*time.Second)
  return err
}
```

Set `skipImage` true (or `DALLESERVER_SKIP_IMAGE=1`) for fast / offline usage.

## .env example

See `.env.example` included in the repo for a documented starter file.

## Implementation notes

Key server concerns illustrated here:

* Per-(series,address) locking + TTL to avoid duplicate work
* Simple context + prompt caching inside the `trueblocks-dalle` library
* Timeouts on enhancement + image requests (configurable)
* Prompt + image phase logging (start/end + elapsed)
* Lint (golangci-lint) pinned version for reproducibility
* Benchmarks + baseline JSON artifacts for regression tracking
* Graceful shutdown and HTTP server timeouts (Slowloris protection)

## Linting & testing

```bash
make lint      # runs golangci-lint
make test      # skips network/image by setting DALLESERVER_SKIP_IMAGE=1 internally
```

Run a single benchmark:

```bash
go test -bench=BenchmarkGenerateAnnotatedImage -run=^$ ./...
```

Capture a baseline JSON (for dashboards / diffing):

```bash
make bench-baseline
```

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Immediate exit: `OPENAI_API_KEY not set` | Missing key & not in skip mode | Export key or set `DALLESERVER_SKIP_IMAGE=1` |
| Enhancement timeout | Model slow / low timeout | Increase `DALLESERVER_ENHANCE_TIMEOUT` |
| Blank preview page | No images yet | Trigger generation (`?generate=1`) |
| 404 under `/files/` | File not generated yet | Wait for generation to complete |

## License

GNU GPL v3 (or later). See `LICENSE`.

## Contributing

PRs welcome. Please see the core project’s [branching workflow](https://github.com/TrueBlocks/trueblocks-core/blob/develop/docs/BRANCHING.md) for consistency.

1. Fork & branch.
2. Make changes + add tests when practical.
3. `make lint test` must pass.
4. Open PR.

## Contact

Questions / ideas: join our Discord (linked from [https://trueblocks.io](https://trueblocks.io)).

## Contributors

Thanks to:

* [@tjayrush](https://github.com/tjayrush)
* [@mikeghen](https://github.com/mikeghen)
* And the broader TrueBlocks community
