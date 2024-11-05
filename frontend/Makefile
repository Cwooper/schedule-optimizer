# Variables
BUILD_DIR := ../build
NODE_MODULES := node_modules
VITE_CONFIG := vite.config.ts

# Default target
.PHONY: all
all: clean build

# Clean build directory and node_modules
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf dist

# Install dependencies
.PHONY: install
install:
	npm install

# Build the project
.PHONY: build
build: install
	npm run build
	# Create build directory if it doesn't exist
	mkdir -p $(BUILD_DIR)
	# Move the built files to the build directory
	cp -r dist/* $(BUILD_DIR)/
	rm -rf dist

# Run development server
.PHONY: run
run: install
	npm run dev

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all     - Clean and build the project (default)"
	@echo "  clean   - Remove build artifacts"
	@echo "  install - Install dependencies"
	@echo "  build   - Build the project"
	@echo "  run     - Run development server"
	@echo "  help    - Show this help message"
	