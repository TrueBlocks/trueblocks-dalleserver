# DalleDress - Better Blockies Through Science

## trueblocks-dalleserver

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

### Clone and run the repo

```[bash]
git clone -b develop git@github.com:TrueBlocks/trueblocks-dalleserver.git
cd dalleserver
go run *.go
```

This should start the dalledress server running on port 8080.

### Requesting images

```[bash]
curl http://localhost:8080/dalle/<series>/<address> --output image.png
```

where `<series>` is one of the results listed with `curl http://localhost:8080/series` and `<address>` is a valid Ethereum address (do not use `.eth` names).

The above `curl` command will either return an image file or a "Pending" request. If pending, revisit the URL with the same parameters to get the image a few seconds later.

### Example

In one terminal window:

```[bash]
go run *.go
```

In another terminal window:

```[bash]
open http://localhost:8080/simple/0xf503017d7baf7fbc0fff7492b751025c6a78179b
```

While the image is being generated, you will be told to return soon. When the image is ready, it will be returned. You can check the response's `Content-Type` to see if the image is ready.

### Listing available series

A `series` is a filter on the databases used to create an image. The simplest series (one with no filter that therefore uses all the databases) is called `simple`. To see all the available series, use the following command:

```[bash]
open http://localhost:8080/series
```

This should return a string list similar to this:

```
Available series:  [
  ...
  "five-tone-postal-protozoa",
  "happy-punk-cats",
  ...
]
```

`five-time-postal-protozoa` limits the databases to `protozoa` who are going `postal` using an artistic style of `five-tone` pencil drawing. `happy-punk-cats` does as you might expect.

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
