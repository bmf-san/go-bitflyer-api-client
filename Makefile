.PHONY: install-tools
install-tools: ## Install tools.
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/lerenn/asyncapi-codegen/cmd/asyncapi-codegen@latest

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run ./...

.PHONY: test
test: ## Run tests.
	go test -v -race ./...

.PHONY: generate
generate: generate-http generate-websocket ## Generate HTTP and WebSocket API client code.

.PHONY: generate-http
generate-http: ## Generate HTTP API client code.
	oapi-codegen -config api/openapi/oapi-codegen-config.yaml api/openapi/http_api.yaml

.PHONY: generate-websocket
generate-websocket: ## Generate WebSocket API client code.
	mkdir -p ./client/websocket
	asyncapi-codegen -i api/asyncapi/realtime_api.yaml -o ./client/websocket/client.gen.go -p websocket

.PHONY: example-http
example-http: ## Run HTTP client example code.
	go run examples/http_client_example.go

.PHONY: example-websocket
example-websocket: ## Run WebSocket client example code.
	cd examples/websocket && go run main.go

.PHONY: example
example: example-http ## Run example code (HTTP client only).

.PHONY: build
build: ## Build the library (verify compilation)
	@echo "Building go-bitflyer-api-client..."
	@go build ./...
	@echo "Build successful"

.PHONY: deps
deps: ## Update dependencies
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download

.PHONY: clean
clean: ## Clean generated files
	@echo "Cleaning generated files..."
	@rm -f client/http/client.gen.go
	@rm -f client/websocket/client.gen.go

.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'