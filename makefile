all:
	go build ./...

run:
	$(MAKE) -j 12 all
	go run .

lint:
	golangci-lint run ./...

test:
	@TB_DALLE_SKIP_IMAGE=1 go test --count=1 ./...
	@cd dalle ; TB_DALLE_SKIP_IMAGE=1 go test --count=1 ./... ; cd - 2>/dev/null

race:
	TB_DALLE_SKIP_IMAGE=1 go test -race ./...

bench:
	TB_DALLE_SKIP_IMAGE=1 go test -bench=. -run=^$ ./...

benchmark:
	TB_DALLE_SKIP_IMAGE=1 go test -bench=BenchmarkGenerateAnnotatedImage -benchmem -run=^$ ./...


bench-baseline:
	@mkdir -p benchmarks
	@ts=$$(date +%Y%m%d%H%M%S); \
	  echo "Running benchmark (timestamp $$ts)..."; \
	  TB_DALLE_SKIP_IMAGE=1 go test -bench=BenchmarkGenerateAnnotatedImage -benchmem -run=^$$ -count=1 -json ./... > benchmarks/$$ts.json; \
	  cp benchmarks/$$ts.json benchmarks/latest.json; \
	  echo "Saved benchmark baseline to benchmarks/$$ts.json and updated benchmarks/latest.json"

# Build & serve documentation book (mdBook) from ./book
.PHONY: book
book:
	$(MAKE) -C book serve

