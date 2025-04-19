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