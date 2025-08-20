# Getting Started

## Prerequisites

- Go (module uses a recent Go toolchain; verify with `go version`)
- An OpenAI API key (required for full image generation)

## Clone

```bash
git clone https://github.com/TrueBlocks/trueblocks-dalleserver.git
cd trueblocks-dalleserver
```

## Configure Environment

Copy the example env file and edit:

```bash
cp .env.example .env
```

Edit `.env` and set:

- `OPENAI_API_KEY` (required unless skipping image generation)
- (Optional) other variables, see [Configuration](./configuration.md)

## Build

```bash
make run   # builds then runs the server
```

If the OpenAI key is missing the server exits with a fatal message.
