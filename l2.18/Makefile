LOCAL_BIN := $(CURDIR)/bin
BASE_STACK := docker compose -f docker-compose.yml
INTEGRATION_TEST_STACK := docker compose --env-file .env -f tests/integration/docker-compose-integration-test.yml
INTEGRATION_TEST_DIR := $(CURDIR)/tests/integration
E2E_TEST_STACK := docker compose --env-file ../../.env -f tests/e2e/docker-compose-e2e.yml
E2E_TEST_DIR := $(CURDIR)/tests/e2e
ALL_STACK := $(BASE_STACK)

.DEFAULT_GOAL := help

.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Tidy and verify Go modules
	go mod tidy && go mod verify

.PHONY: deps-audit
deps-audit: ## Check dependencies for vulnerabilities using govulncheck (govulncheck is must be required)
	govulncheck ./...

.PHONY: bin-deps
bin-deps: ## Install development tools (govulncheck, golangci-lint, gci, gofumpt, etc.)
	@echo "Installing development tools..."
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
	go install github.com/daixiang0/gci@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/segmentio/golines@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/vektra/mockery/v3@v3.5.5
	@echo "Development tools installed."

.PHONY: format
format: ## Format code using gofumpt, gci, golines, goimports
	@echo "Formatting code..."
	gofumpt -l -w .
	gci write . --skip-generated -s standard -s default
	golines -w --max-len=120 .
	goimports -w .
	@echo "Code formatted."

.PHONY: linter-golangci
linter-golangci: ## Run golangci-lint linter
	golangci-lint run

.PHONY: linter-hadolint
linter-hadolint: ## Run hadolint on Dockerfiles (requires hadolint installed)
	hadolint Dockerfile
	hadolint tests/integration/Dockerfile

.PHONY: linter-dotenv
linter-dotenv: ## Run dotenv-linter on .env files (requires dotenv-linter installed)
	dotenv-linter -r || echo "dotenv-linter not found or issues found in .env files"

.PHONY: swag-v1
swag-v1: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	swag init -g internal/transport/routes.go --output docs
	@echo "Swagger documentation generated."

.PHONY: mock
mock: ## Generate mocks using mockgen
	@echo "Generating mocks..."
	mockgen -source ./internal/service/service.go -destination ./internal/repository/mock/repository.go -package=mock_repository
	mockgen -source ./pkg/cache/cache.go -destination ./pkg/cache/mock/cache.go -package=mock_cache
	mockgen -source ./pkg/logger/logger.go -destination ./pkg/logger/mock/logger.go -package=mock_logger
	@echo "Mocks generated."

.PHONY: run
run: deps swag-v1 ## Run the application locally
	@echo "Running application..."
	go run ./cmd/calendar-service -config=./configs/dev.env

.PHONY: compose-up
compose-up-all: ## Run calendar-service in container
	$(BASE_STACK) up --build -d
	$(BASE_STACK) logs -f

.PHONY: compose-down
compose-down: ## Stop and remove all containers, networks, and volumes (from all stacks)
	$(ALL_STACK) down --remove-orphans --volumes

.PHONY: compose-logs
compose-logs: ## Follow logs for service
	$(BASE_STACK) logs -f

.PHONY: test
test: ## Run unit tests with race detector and coverage
	@echo "Running unit tests..."
	go clean -testcache
	go test -v -race -covermode atomic -coverprofile=coverage_internal.txt ./internal/...
	go test -v -race -covermode atomic -coverprofile=coverage_pkg.txt ./pkg/...
	@echo "Unit tests completed."

.PHONY: integration-test
integration-test: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	$(INTEGRATION_TEST_STACK) up db -d
	$(INTEGRATION_TEST_STACK) run --rm db-migrator
	$(INTEGRATION_TEST_STACK) up integration-test --exit-code-from integration-test
	$(INTEGRATION_TEST_STACK) down --remove-orphans --volumes
	@echo "Integration tests completed."

.PHONY: pre-commit
pre-commit: swag-v1 mock format linter-golangci test ## Run checks typically done before committing
	@echo "Pre-commit checks passed."

.PHONY: build
build: deps ## Build the main application binary
	@echo "Building application binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/calendar-service ./cmd/calendar-service
	@echo "Binary built: ./bin/calendar-service"

.PHONY: build-docker
build-docker: ## Build main Docker image
	@echo "Building main Docker image..."
	docker build -t calendar-service:latest .
	@echo "Main Docker image built."

.PHONY: clean
clean: ## Remove generated files and binaries
	@echo "Cleaning up..."
	rm -rf ./bin/
	rm -rf ./docs/ # Swagger docs
	find . -name "*mock*" -type f -path "*/mock/*" -delete
	@echo "Cleanup completed."

.PHONY: docker-prune
docker-prune: ## Remove unused Docker data (stopped containers, networks, images, build cache)
	@echo "Pruning Docker data..."
	docker system prune -af
	@echo "Docker data pruned."
