all:
	go build ./...

serve:
	@make test
	@$(MAKE) -j 12 all
	@go run .

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

