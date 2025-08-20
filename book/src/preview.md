# Preview Gallery

`/preview` renders an HTML page listing annotated images grouped by series.

Implementation outline (`handlePreview`):
1. Walk `output/` recursively.
2. Select PNG files whose path includes `/annotated/`.
3. Extract series (first path segment) and address (filename sans extension).
4. Collect modification time and relative path.
5. Group by series, sort each group by descending modification time.
6. Render a Go `html/template` with a client-side JavaScript filter.

Each figure shows:
- Image (square container with `object-fit: contain`)
- Address (EIP-55 case)
- Timestamp (mod time formatted `YYYY-MM-DD HH:MM:SS`)
