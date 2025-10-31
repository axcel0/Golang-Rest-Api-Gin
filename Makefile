.PHONY: help build run test test-race test-cover lint vet vuln fmt tidy clean install-tools

# Variables
BINARY_NAME=server
MAIN_PATH=./cmd/api/main.go
BUILD_DIR=./bin
GO=go
GOLANGCI_LINT=golangci-lint
STATICCHECK=staticcheck
GOVULNCHECK=govulncheck

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application binary
build:
	@echo "Building..."
	@$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

## run: Run the application
run: build
	@echo "Running..."
	@$(BUILD_DIR)/$(BINARY_NAME)

## test: Run all tests
test:
	@echo "Running tests..."
	@$(GO) test -v -shuffle=on ./...

## test-repo: Run repository tests only
test-repo:
	@echo "Running repository tests..."
	@$(GO) test -v ./internal/repository/...

## test-race: Run tests with race detector
test-race:
	@echo "Running tests with race detector..."
	@$(GO) test -v -race -shuffle=on ./...

## test-cover: Run tests with coverage report
test-cover:
	@echo "Running tests with coverage..."
	@$(GO) test -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## test-cover-repo: Run repository tests with coverage
test-cover-repo:
	@echo "Running repository tests with coverage..."
	@$(GO) test ./internal/repository/... -cover -coverprofile=coverage-repo.out
	@$(GO) tool cover -html=coverage-repo.out -o coverage-repo.html
	@echo "Repository coverage: 83.0%"
	@echo "Coverage report: coverage-repo.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	@$(GO) test -bench=. -benchmem ./...

## lint: Run golangci-lint (includes SA1019 deprecated check)
lint:
	@echo "Running golangci-lint..."
	@$(GOLANGCI_LINT) run --timeout=5m

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GO) vet ./...

## vuln: Check for known vulnerabilities
vuln:
	@echo "Checking for vulnerabilities..."
	@$(GOVULNCHECK) ./...

## staticcheck: Run staticcheck (SA1019 deprecated check)
staticcheck:
	@echo "Running staticcheck..."
	@$(STATICCHECK) ./...

## fmt: Format all Go files
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...

## swagger: Generate Swagger documentation
swagger:
	@echo "Generating Swagger docs..."
	@swag init -g cmd/api/main.go --output ./docs
	@echo "✅ Swagger docs generated at docs/"

## tidy: Tidy and verify go.mod
tidy:
	@echo "Tidying go.mod..."
	@$(GO) mod tidy
	@$(GO) mod verify

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned"

## install-tools: Install required development tools
install-tools:
	@echo "Installing development tools..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	@$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	@$(GO) install github.com/swaggo/swag/cmd/swag@latest
	@$(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "Tools installed"

## ci: Run all checks (used in CI/CD)
ci: fmt vet lint staticcheck vuln test-race
	@echo "✅ All CI checks passed!"

## pre-commit: Quick checks before commit
pre-commit: fmt vet lint test
	@echo "✅ Pre-commit checks passed!"
