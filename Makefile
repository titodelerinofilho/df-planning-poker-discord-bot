GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOCACHE ?= $(CURDIR)/.cache/go-build
GOLANGCI_LINT_CACHE ?= $(CURDIR)/.cache/golangci-lint
BIN_DIR ?= bin
BOT_BIN ?= $(BIN_DIR)/bot
DOCKER_IMAGE ?= df-planning-poker-discord-bot
BOT_PACKAGE := ./cmd/bot
PACKAGES := ./...

.PHONY: run test test-race vet lint build docker-build clean

run:
	GOCACHE=$(GOCACHE) $(GO) run $(BOT_PACKAGE)

test:
	GOCACHE=$(GOCACHE) $(GO) test $(PACKAGES)

test-race:
	GOCACHE=$(GOCACHE) $(GO) test -race $(PACKAGES)

vet:
	GOCACHE=$(GOCACHE) $(GO) vet $(PACKAGES)

lint:
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		GOCACHE=$(GOCACHE) GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) $(GOLANGCI_LINT) run; \
	else \
		echo "golangci-lint is not installed. Install it from https://golangci-lint.run/welcome/install/"; \
		exit 127; \
	fi

build:
	mkdir -p $(BIN_DIR)
	GOCACHE=$(GOCACHE) $(GO) build -o $(BOT_BIN) $(BOT_PACKAGE)

docker-build:
	docker build -t $(DOCKER_IMAGE) .

clean:
	rm -rf $(BIN_DIR) .cache
