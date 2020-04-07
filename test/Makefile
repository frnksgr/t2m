.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; \
		{printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## test haproxy in front of multiple t2m backends
	docker-compose up --scale backend=5 -d
	for t in binary linear fan; do \
		for c in 10 100 1000; do \
			curl "http://localhost:8080?size=$$c&topology=$$t" | jq .; \
		done; \
	done
	docker-compose down

up: ## bring up haproxy in front of multiple t2m backends NOTE: runs interactively
	@echo send requests to http://localhost:8080
	docker-compose up --scale backend=5
