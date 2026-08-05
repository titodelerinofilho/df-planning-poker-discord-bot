GO ?= go
GOCACHE ?= $(CURDIR)/.cache/go-build
BIN_DIR ?= bin
BOT_BIN ?= $(BIN_DIR)/bot
BOT_PACKAGE := ./cmd/bot
PACKAGES := ./...

.PHONY: run test test-race vet lint build clean

run:
	GOCACHE=$(GOCACHE) $(GO) run $(BOT_PACKAGE)

test:
	GOCACHE=$(GOCACHE) $(GO) test $(PACKAGES)

test-race:
	GOCACHE=$(GOCACHE) $(GO) test -race $(PACKAGES)

vet:
	GOCACHE=$(GOCACHE) $(GO) vet $(PACKAGES)

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed yet; lint configuration is planned in issue 0.2.1"; \
	fi

build:
	mkdir -p $(BIN_DIR)
	GOCACHE=$(GOCACHE) $(GO) build -o $(BOT_BIN) $(BOT_PACKAGE)

clean:
	rm -rf $(BIN_DIR) .cache
