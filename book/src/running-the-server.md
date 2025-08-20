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
- Override via `--port=<n>` command-line flag or `DALLESERVER_PORT` env var.
