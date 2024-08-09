# DalleServer - Better Blockies Through Science

DalleServer is an API server that creates images of Ethereum addresses using the
[DALL-E](https://openai.com/research/dall-e/) model. The server is written in Go
and can be accessed via a RESTful API.

<img alt="google" src="https://www.google.com/logos/doodles/2024/paris-games-breaking-6753651837110566-law.gif">

## Installing

### Install latest [GoLang](https://go.dev/doc/install)

```[bash]
➤ go version
go version go1.22.0 darwin/arm64
```

### Install latest [Wails](https://wails.io/docs/gettingstarted/installation/)

```[bash]
➤ wails version
v2.8.2
```

### Clone and run the server

```[bash]
git clone -b develop git@github.com:TrueBlocks/trueblocks-dalleserver.git
cd dalleserver
yarn serve
```

This will start the API server at `http://localhost:8080`. (If port :8080 is in use, add `--port=<n>` to the `yarn serve` command.)

### Requesting an image

Open your browser window to:

```[bash]
open http://localhost:8080/dalle/<series>/<address>
```

where `<series>` is one of the results listed with `curl http://localhost:8080/series` and `<address>` is a valid Ethereum address (do not use `.eth` names).

The above `curl` command will either return an image file or message telling you to return shortly. You may revisit the URL repeatedly until the image appears.

### Example

In one terminal window:

```[bash]
yarn serve
```

```[bash]
Open your browser to:

```[bash]
open http://localhost:8080/simple/0xf503017d7baf7fbc0fff7492b751025c6a78179b
```

While the image is being generated, you will get a message telling you to return. If you're accessing the API programatically, you may check the response's `Content-Type` to see if the image is ready.

### Listing available series

A `series` is a filter on the databases used to create an image. The simplest series is empty and is called `simple`. To see all the available series, use the following command:

```[bash]
open http://localhost:8080/series
```

This should return a string list similar to this:

```[bash]
Available series:  [
  ...
  "five-tone-postal-protozoa",
  "happy-punk-cats",
  ...
]
```

`five-time-postal-protozoa` limits the databases to `protozoa` who are going `postal` using an artistic style of `five-tone` pencil drawing. `happy-punk-cats` does as you might expect.

You may see the details of the filter by appending `?details=<series>` to the above URL.

## Contributing

We love contributors. Please see information about our [workflow](https://github.com/TrueBlocks/trueblocks-core/blob/develop/docs/BRANCHING.md) before proceeding.

1. Fork this repository into your own repo.
2. Create a branch: `git checkout -b <branch_name>`.
3. Make changes to your local branch and commit them to your forked repo: `git commit -m '<commit_message>'`
4. Push back to the original branch: `git push origin TrueBlocks/trueblocks-core`
5. Create the pull request.

## Contact

If you have questions, comments, or complaints, please join the discussion on our discord server which is [linked from our website](https://trueblocks.io).

## List of Contributors

Thanks to the following people who have contributed to this project:

- [@tjayrush](https://github.com/tjayrush)
- [@mikeghen](https://github.com/mikeghen)
- many others
