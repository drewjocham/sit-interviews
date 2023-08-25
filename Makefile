PROJ_PATH=${CURDIR}
.DEFAULT_GOAL := help

.PHONY: api
api: ## run the api
	cd api && go run ./cmd

.PHONY: build-api
build-api: ## build api for production
	cd api/cmd/api && go build

.PHONY: docker-build
docker-build: ## start docker compose
	docker-compose build

.PHONY: up
up: ## start docker compose
	docker-compose up -d

.PHONY: down
down: ## start docker compose
	docker-compose down

.PHONY: mod
mod: ## Download, verify and vendor dependencies
	cd api && go mod tidy && go mod download && go mod verify

.PHONY: linter-fix
linter-fix: ## Fix linter api errors
	cd api && golangci-lint run --fix

.PHONY: linter
linter: ## Run linter
	cd api && golangci-lint run

.PHONY: update-go-deps
update-go-deps:
	cd api && go get -u ./...

.PHONY: help
help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''
