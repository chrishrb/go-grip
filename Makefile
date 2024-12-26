.PHONY: all run format emojiscraper build vendor test compile format lint clean

# If the first argument is "run"...
ifeq (run,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

GOCMD=go
LDFLAGS="-s -w ${LDFLAGS_OPT}"

all: vendor build format lint ## Format, lint and build

run: ## Run
	go run -tags debug main.go $(RUN_ARGS)

emojiscraper: ## Run emojiscraper
	go run -tags debug main.go emojiscraper defaults/static/emojis pkg/emoji_map.go

build: ## Build
	go build -tags debug -o bin/go-grip main.go

vendor: ## Vendor
	go mod vendor

test: ## Test
	${GOCMD} test ./...

compile: ## Compile for every OS and Platform
	echo "Compiling for every OS and Platform"
	GOOS=darwin GOARCH=amd64 go build -o bin/go-grip-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/go-grip-darwin-arm64 main.go
	GOOS=linux GOARCH=amd64 go build -o bin/go-grip-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -o bin/go-grip-linux-arm64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/go-grip-windows-amd64.exe main.go
	GOOS=windows GOARCH=arm64 go build -o bin/go-grip-windows-arm64.exe main.go

format: ## Format code
	${GOCMD} fmt ./...

lint: ## Run linter
	golangci-lint run

clean: ## Cleanup build dir
	rm -r bin/

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
