.PHONY: install run clean dev test lint format help

# Default target
.DEFAULT_GOAL := help

# Install dependencies
install:
	@echo "Installing dependencies..."
	npm install

# Run the example
run:
	@echo "Running JavaScript client example..."
	npm start

# Run in development mode (with nodemon if installed)
dev:
	@echo "Running in development mode..."
	@if command -v nodemon >/dev/null 2>&1; then \
		nodemon index.js; \
	else \
		echo "nodemon not found, running with node..."; \
		npm start; \
	fi

# Clean node_modules and package-lock.json
clean:
	@echo "Cleaning up..."
	rm -rf node_modules
	rm -f package-lock.json

# Reinstall dependencies
reinstall: clean install
	@echo "Dependencies reinstalled"

# Run tests (when available)
test:
	@echo "Running tests..."
	npm test

# Lint code (if linter is configured)
lint:
	@echo "Linting code..."
	@if command -v eslint >/dev/null 2>&1; then \
		eslint *.js; \
	else \
		echo "ESLint not found, skipping lint..."; \
	fi

# Format code (if prettier is available)
format:
	@echo "Formatting code..."
	@if command -v prettier >/dev/null 2>&1; then \
		prettier --write *.js; \
	else \
		echo "Prettier not found, skipping format..."; \
	fi

# Install development tools
install-dev:
	@echo "Installing development tools..."
	npm install --save-dev nodemon eslint prettier

# Check Node.js version
check:
	@echo "Checking Node.js version..."
	@node --version
	@echo "Checking npm version..."
	@npm --version

# Show help
help:
	@echo "Available commands:"
	@echo "  install     - Install dependencies"
	@echo "  run         - Run the JavaScript client example"
	@echo "  dev         - Run in development mode (with nodemon)"
	@echo "  clean       - Remove node_modules and package-lock.json"
	@echo "  reinstall   - Clean and reinstall dependencies"
	@echo "  test        - Run tests"
	@echo "  lint        - Lint code with ESLint"
	@echo "  format      - Format code with Prettier" 
	@echo "  install-dev - Install development tools (nodemon, eslint, prettier)"
	@echo "  check       - Check Node.js and npm versions"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  make install && make run"
