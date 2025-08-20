# Introduction

`trueblocks-dalleserver` is a Go HTTP server that generates annotated images for Ethereum addresses.

It demonstrates how to:

- List available image generation "series"
- Generate prompts and images for a specific address
- Serve generated PNGs and prompt artifacts from an organized `output/` directory
- Provide a simple HTML preview gallery of annotated images
- Expose health and metrics endpoints

The server relies on an external library (`trueblocks-dalle`) for prompt & image generation logic. This book documents only the code present in this repository (excluding the submodule internals).
