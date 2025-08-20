SRC_MD := $(shell find src -name '*.md')
SRC_PNG := $(shell find src -name '*.png')

book: $(SRC_MD) $(SRC_PNG) book.toml makefile
	@echo building mdBook...
	@mdbook build

serve:
	@$(MAKE) clean
	@$(MAKE) book
	@mdbook serve --open --hostname 127.0.0.1

clean:
	@rm -rf book/html book/book.toml.lock || true

# Optional deploy target (adjust destination before use)
# deploy:
# 	@$(MAKE) clean && touch ${SRC_MD} && $(MAKE) book && cd book && scp -pr * user@host:/path/to/deploy/
