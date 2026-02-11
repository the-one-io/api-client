APP_NAME := client
SRC := main.go
BUILD_DIR := build

GO_FILES := $(shell find . -type f -name '*.go')

.PHONY: build run run-ws deps clean build-all build-windows build-macos build-linux

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod tidy

run: deps ## Run the REST API client example
	@echo "🚀 Running REST API client example..."
	@CGO_ENABLED=0 go build -o ./broker-rest-client main.go http_client.go
	@./broker-rest-client
	@rm -f ./broker-rest-client

run-ws: deps ## Run the WebSocket client example
	@echo "🌐 Running WebSocket client example..."
	@CGO_ENABLED=0 go build -o ./broker-ws-client ws_main.go ws_client.go ws_example.go
	@./broker-ws-client
	@rm -f ./broker-ws-client

clean:
	@echo "🧹 Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@mkdir -p $(BUILD_DIR)

build-all: build-windows build-macos build-linux
	@echo "✅ All builds completed."

build-windows: $(GO_FILES)
	@echo "🪟 Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-win.exe $(SRC)

build-macos: $(GO_FILES)
	@echo "🍎 Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-mac $(SRC)
	@echo "🍏 Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-mac-arm $(SRC)

build-linux: $(GO_FILES)
	@echo "🐧 Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux $(SRC)

.DEFAULT_GOAL := run
