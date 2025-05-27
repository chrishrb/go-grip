.PHONY: all run format emojiscraper build vendor test compile format lint clean release

# If the first argument is "run"...
ifeq (run,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

GOCMD=go

# pkgver variables
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-s -w
VERSION_FLAGS=-X github.com/chrishrb/go-grip/cmd.Version=$(VERSION) \
	      -X github.com/chrishrb/go-grip/cmd.CommitHash=$(COMMIT) \
	      -X github.com/chrishrb/go-grip/cmd.BuildDate=$(BUILD_DATE) \
	      FULL_LDFLAGS=-ldflags "$(LDFLAGS) $(VERSION_FLAGS)"

all: vendor build format lint ## Format, lint and build

run: ## Run
	go run -tags debug main.go $(RUN_ARGS)

emojiscraper: ## Run emojiscraper
	go run -tags debug main.go emojiscraper defaults/static/emojis pkg/emoji_map.go

build: ## Build
	$(GOCMD) build -tags debug $(FULL_LDFLAGS) -o bin/go-grip main.go

vendor: ## Vendor
	$(GOCMD) mod vendor

test: ## Test
	$(GOCMD) test ./...

compile: ## Compile for every OS and Platform
	echo "Compiling for every OS and Platform"
	GOOS=darwin GOARCH=amd64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/go-grip-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/go-grip-darwin-arm64 main.go
	GOOS=linux GOARCH=amd64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/go-grip-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/go-grip-linux-arm64 main.go
	GOOS=windows GOARCH=amd64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/go-grip-windows-amd64.exe main.go
	GOOS=windows GOARCH=arm64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/go-grip-windows-arm64.exe main.go

release: clean
	@echo "Creating release $(VERSION)"
	@if echo $(VERSION) | grep -q "dirty"; then \
		echo "Error: dirty repo."; \
	fi
	mkdir -p bin/release
	GOOS=darwin GOARCH=amd64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/release/go-grip-$(VERSION)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/release/go-grip-$(VERSION)-darwin-arm64 main.go
	GOOS=linux GOARCH=amd64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/release/go-grip-$(VERSION)-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/release/go-grip-$(VERSION)-linux-arm64 main.go
	GOOS=windows GOARCH=amd64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/release/go-grip-$(VERSION)-windows-amd64.exe main.go
	GOOS=windows GOARCH=arm64 $(GOCMD) build $(FULL_LDFLAGS) -o bin/release/go-grip-$(VERSION)-windows-arm64.exe main.go
	@echo "Release $(VERSION) created in bin/release/"

format: ## Format code
	$(GOCMD) fmt ./...

lint: ## Run linter
	golangci-lint run

clean: ## Cleanup build dir
	rm -rf bin/
	@go mod tidy

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
