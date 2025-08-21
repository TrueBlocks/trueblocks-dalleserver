# DalleDress Schema

The `DalleDress` object is embedded in every progress snapshot and represents both the *inputs* (attributes/seed) and *derived prompts* plus final artifact metadata.

All fields are always present in JSON (no `omitempty`). Empty slices serialize as `[]`.

| Field | Type | Description |
|-------|------|-------------|
| `original` | string | Original address / identifier used to create the dress. |
| `fileName` | string | Sanitized deterministic filename stem for outputs. |
| `seed` | string | 64 hex chars (derived, truncated) used to drive attribute selection. |
| `prompt` | string | The full base prompt (template expansion). |
| `dataPrompt` | string | Data-oriented listing of selected attributes. |
| `titlePrompt` | string | Concise title version. |
| `tersePrompt` | string | Short prompt form (adjectives + noun + emotion, etc.). |
| `enhancedPrompt` | string | LLM-enhanced prompt (may equal base if enhancement skipped). |
| `attributes` | Attribute[] | Ordered list of generated attributes (struct list). |
| `seedChunks` | string[] | Extracted segments of the seed used in attribute derivation. |
| `selectedTokens` | string[] | Canonical attribute names/token identifiers selected. |
| `selectedRecords` | string[] | Values (or value fragments) associated with tokens. |
| `imageUrl` | string | Remote image URL returned by DALL·E (set after request). |
| `annotatedPath` | string | Local filesystem path to annotated PNG (set after annotate phase). |
| `ipfsHash` | string | Reserved for future IPFS pin result (currently blank). |
| `cacheHit` | bool | True if this dress run was satisfied from an existing annotated image. |
| `completed` | bool | True when generation finished (success, error, or cache). |

`AttribMap` exists in Go but is intentionally excluded from JSON (internal lookup map).

## Lifecycle Population
1. Construction (`MakeDalleDress`) fills: original, fileName, seed, *prompts*, attributes, seedChunks, selectedTokens, selectedRecords.
2. Enhancement phase may update `enhancedPrompt` (or leave as empty/base if skipped).
3. Image request sets `imageUrl` when the API responds.
4. Annotation phase sets `annotatedPath`.
5. Completion sets `completed=true` and (if applicable) `cacheHit=true` (on fast path).

## Stability & Backward Compatibility
Field names are stable; new fields (e.g. future `ipfsHash`) will be appended rather than renamed. Clients should ignore unknown fields for forward compatibility.

## Example (truncated)
```json
{
  "original": "0xf50301...",
  "fileName": "0xf50301...",
  "seed": "f503017d7baf7fbc0fff7492b751025c6a78179b...",
  "prompt": "Draw a ...",
  "dataPrompt": "Adverb: ...",
  "titlePrompt": "Joyful ...",
  "tersePrompt": "vivid ...",
  "enhancedPrompt": "DO NOT PUT TEXT ...",
  "attributes": [ {"name":"adverb", "value":"..."}, ... ],
  "seedChunks": ["abcd12", "ef3456"],
  "selectedTokens": ["adverb", "adjective", "noun"],
  "selectedRecords": ["fast, quickly", "bright, luminous"],
  "imageUrl": "https://.../image.png",
  "annotatedPath": "/data/output/simple/annotated/0xf50301....png",
  "ipfsHash": "",
  "cacheHit": false,
  "completed": true
}
```
