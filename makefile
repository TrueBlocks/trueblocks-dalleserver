BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
BUILD_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
VERSION := $(shell git describe --tags --exact-match 2>/dev/null || echo "development")

LDFLAGS := -X 'main.BuildTime=$(BUILD_TIME)' \
		   -X 'main.BuildCommit=$(BUILD_COMMIT)' \
		   -X 'main.BuildBranch=$(BUILD_BRANCH)' \
		   -X 'main.Version=$(VERSION)'

all:
	go build -ldflags "$(LDFLAGS)" ./...

build:
	go build -ldflags "$(LDFLAGS)" -o trueblocks-dalleserver .

serve:
	@make test
	@$(MAKE) -j 12 all
	@go run -ldflags "$(LDFLAGS)" .

lint:
	golangci-lint run ./...

test:
	@TB_DALLE_SKIP_IMAGE=1 go test ./...

build-db:
	@cd dalle ; make build-db ; cd - 2>/dev/null

race:
	TB_DALLE_SKIP_IMAGE=1 go test -race ./...

bench:
	TB_DALLE_SKIP_IMAGE=1 go test -bench=. -run=^$ ./...

benchmark:
	TB_DALLE_SKIP_IMAGE=1 go test -bench=BenchmarkGenerateAnnotatedImage -benchmem -run=^$ ./...


# Build & serve documentation book (mdBook) from ./book
.PHONY: book
book:
	$(MAKE) -C book serve

# Build documentation book (mdBook) without serving
.PHONY: build-book
build-book:
	$(MAKE) -C book book

