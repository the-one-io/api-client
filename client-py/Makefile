.PHONY: install run clean venv test lint format help

# Default target
.DEFAULT_GOAL := help

# Python executable
PYTHON := python3
PIP := pip3

# Virtual environment directory
VENV_DIR := venv

# Install dependencies
install:
	@echo "Installing dependencies..."
	$(PIP) install -r requirements.txt

# Install dependencies with virtual environment
install-venv: venv
	@echo "Installing dependencies in virtual environment..."
	./$(VENV_DIR)/bin/pip install -r requirements.txt

# Create virtual environment
venv:
	@echo "Creating virtual environment..."
	$(PYTHON) -m venv $(VENV_DIR)

# Run the example
run:
	@echo "Running Python client example..."
	$(PYTHON) example.py

# Run with virtual environment
run-venv: venv install-venv
	@echo "Running Python client example in virtual environment..."
	./$(VENV_DIR)/bin/python example.py

# Clean virtual environment and cache
clean:
	@echo "Cleaning up..."
	rm -rf $(VENV_DIR)
	rm -rf __pycache__
	rm -rf *.pyc
	rm -rf .pytest_cache
	find . -name "*.pyc" -delete
	find . -name "__pycache__" -delete

# Run tests (when available)
test:
	@echo "Running tests..."
	@if command -v pytest >/dev/null 2>&1; then \
		pytest; \
	else \
		echo "pytest not found, install with: pip install pytest"; \
	fi

# Lint code
lint:
	@echo "Linting code..."
	@if command -v flake8 >/dev/null 2>&1; then \
		flake8 *.py; \
	else \
		echo "flake8 not found, install with: pip install flake8"; \
	fi

# Format code
format:
	@echo "Formatting code..."
	@if command -v black >/dev/null 2>&1; then \
		black *.py; \
	else \
		echo "black not found, install with: pip install black"; \
	fi

# Type checking
typecheck:
	@echo "Type checking..."
	@if command -v mypy >/dev/null 2>&1; then \
		mypy *.py; \
	else \
		echo "mypy not found, install with: pip install mypy"; \
	fi

# Install development tools
install-dev:
	@echo "Installing development tools..."
	$(PIP) install pytest flake8 black mypy

# Install development tools in virtual environment
install-dev-venv: venv
	@echo "Installing development tools in virtual environment..."
	./$(VENV_DIR)/bin/pip install pytest flake8 black mypy

# Check Python version
check:
	@echo "Checking Python version..."
	@$(PYTHON) --version
	@echo "Checking pip version..."
	@$(PIP) --version

# Freeze requirements
freeze:
	@echo "Freezing requirements..."
	$(PIP) freeze > requirements-frozen.txt

# Show help
help:
	@echo "Available commands:"
	@echo "  install     - Install dependencies"
	@echo "  install-venv - Install dependencies in virtual environment"
	@echo "  venv        - Create virtual environment"
	@echo "  run         - Run the Python client example"
	@echo "  run-venv    - Run the example in virtual environment"
	@echo "  clean       - Remove virtual environment and cache files"
	@echo "  test        - Run tests with pytest"
	@echo "  lint        - Lint code with flake8"
	@echo "  format      - Format code with black"
	@echo "  typecheck   - Type check with mypy"
	@echo "  install-dev - Install development tools"
	@echo "  install-dev-venv - Install development tools in venv"
	@echo "  check       - Check Python and pip versions"
	@echo "  freeze      - Freeze current requirements"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  make install && make run"
	@echo ""
	@echo "With virtual environment:"
	@echo "  make run-venv"
