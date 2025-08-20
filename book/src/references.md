# References

- Main entrypoint: `main.go`
- Configuration loader: `config.go`
- Handlers:
  - Root/endpoints listing: `handle_default.go`
  - Series listing: `handle_series.go`
  - Image request/remove: `handle_dalle.go`
  - Preview gallery: `handle_preview.go`
- Tests:
  - Request parsing & series: `request_test.go`
  - Failure simulation: `failure_test.go`

The image / prompt logic is delegated to the external `trueblocks-dalle` package (not documented here per project scope).

Generated artifacts live under `output/`.
