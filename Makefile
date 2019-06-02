
STATIC ?=
ifdef STATIC
	GOENV := GOOS=linux GORACH=amd64 CGO_ENABLED=0
else
	GOENV := GOOS=linux GORACH=amd64
endif

.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build local to create bin/... executable(s)
	$(GOENV) go build -v ./internal/pkg/...
	$(GOENV) go build -o ./bin/server ./cmd/server/...

.PHONY: clean
clean: ## clean up
	go clean -i ./...
	rm -f bin/server
