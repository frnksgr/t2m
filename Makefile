
STATIC ?=
ifdef STATIC
	GOENV := GOOS=linux GORACH=amd64 CGO_ENABLED=0
else
	GOENV := GOOS=linux GORACH=amd64
endif

DEBUG ?=
ifdef DEBUG
	GOBUILDFLAGS := -gcflags '-m -m'
endif


IMAGE = docker.io/frnksgr/t2m


.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; \
		{printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build local to create bin/... executable(s)
	$(GOENV) go build $(GOBUILDFLAGS) -v ./pkg/...
	$(GOENV) go build $(GOBUILDFLAGS) -o ./bin/server ./cmd/server/...

.PHONY: run
run: build ## run t2m a single process
	@bin/server

.PHONY: clean
clean: ## clean up
	go clean -i ./...
	rm -f bin/server

.PHONY: docker-build
docker-build: ## build docker image
	docker build -t $(IMAGE) .
	docker tag $(IMAGE) $(IMAGE):scratch
	docker build -t $(IMAGE):alpine3.9 --build-arg BASEIMAGE=alpine:3.9 .


.PHONY: docker-push
docker-push: docker-build ## push docker-image
	docker push $(IMAGE)
	docker push $(IMAGE):scratch
	docker push $(IMAGE):alpine3.9
