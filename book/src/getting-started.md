# Getting Started

This section launches the server locally using only the components in this repository. Prompt generation, attribute derivation and series logic reside in the separate `trueblocks-dalle` library—refer to its book for deeper internals.

## Prerequisites

* Go 1.21+ (check with `go version`)
* (Optional) An OpenAI API key for real enhancement + image generation. Without it the server automatically switches to mock/skip mode (still exercising progress + caching paths).

## Clone & Enter

```bash
git clone https://github.com/TrueBlocks/trueblocks-dalleserver.git
cd trueblocks-dalleserver
```

## Environment Setup

Copy and edit an env file:

```bash
cp .env.example .env
```

Populate at minimum (fish shell example):

```fish
set -x OPENAI_API_KEY "sk-..."  # optional; omit to run in skip image mode
```

You can also export at runtime or rely on your shell profile. The minimal config for a mock run is literally nothing—absence of `OPENAI_API_KEY` implies skip image.

## Run

```bash
make run
```

Output (abridged):

```
[status] reporter started (interval=2s)
Starting server on :8080
```

Visit:

* List series: <http://localhost:8080/series>
* Trigger generation: <http://localhost:8080/dalle/simple/0xf503017d7baf7fbc0fff7492b751025c6a78179b?generate=1>
* Poll progress (repeat same URL; `done=true` when finished)
* Gallery: <http://localhost:8080/preview>

## Forcing Regeneration

Remove a cached annotated image by adding `?remove=1` then re‑issuing `?generate=1`.

## Graceful Shutdown

Ctrl+C sends SIGINT. The server initiates a 10s timeout graceful shutdown (`http.Server.Shutdown`) after which it force closes.

## Next

Proceed to [Running the Server](./running-the-server.md) for flags, port overrides and data directory details.
