# Running the Server

Start the server (build + run):

```bash
make run
```

You should see output similar to:

```
Starting server on :8080
```

The server refuses to start if `OPENAI_API_KEY` is not set (unless mock skipping is in effect from environment— see code in `config.go`).

Graceful shutdown: Ctrl+C sends SIGINT which triggers a 10s timeout shutdown sequence.

## Ports

- Default: `:8080`
- Override via `--port=<n>` command-line flag or `TB_DALLE_PORT` env var.

## Data Directory

Runtime artifacts (logs, series definitions, generated output) reside under a base data directory resolved in this order:

1. `--data-dir` flag
2. `TB_DALLE_DATA_DIR` environment variable
3. `$XDG_DATA_HOME/trueblocks/dalle` if `XDG_DATA_HOME` is set
4. `$HOME/.local/share/trueblocks/dalle` fallback

If the chosen directory is not writable the server attempts a temporary fallback and logs a warning.
