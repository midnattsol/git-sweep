.PHONY: build install clean test lint release-local

BINARY_NAME := git-sweep
INSTALL_DIR := $(HOME)/.local/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build binary
build:
	go build -ldflags "-s -w -X github.com/midnattsol/git-sweep/cmd.version=$(VERSION)" -o $(BINARY_NAME) .

# Install locally
install: build
	mkdir -p $(INSTALL_DIR)
	mv $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	git config --global alias.sweep '!git-sweep'
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)"
	@echo "Created git alias: git sweep"

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Run tests
test:
	go test -v ./...

# Lint
lint:
	golangci-lint run

# Test goreleaser locally (dry-run)
release-local:
	goreleaser release --snapshot --clean

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build binary"
	@echo "  install       - Build and install to ~/.local/bin"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linter"
	@echo "  release-local - Test goreleaser locally"
