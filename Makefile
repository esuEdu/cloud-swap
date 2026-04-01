.PHONY: build test clean install run help

BINARY_NAME=cloud-swap
VERSION?=dev
INSTALL_DIR=$(HOME)/.local/bin

build:
	@echo "Building cloud-swap..."
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY_NAME) ./cmd

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	go clean

install: build
	@echo "Installing to $(INSTALL_DIR)..."
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY_NAME) $(INSTALL_DIR)/
	chmod +x $(INSTALL_DIR)/$(BINARY_NAME)

uninstall:
	@echo "Uninstalling..."
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)

run:
	go run ./cmd

tui:
	go run ./cmd tui

deps:
	go mod download
	go mod tidy

lint:
	golangci-lint run ./... || go vet ./...

all: deps build

help:
	@echo "cloud-swap Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build     - Build the binary"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  install   - Build and install binary"
	@echo "  uninstall - Remove installed binary"
	@echo "  run       - Run directly"
	@echo "  tui       - Launch TUI"
	@echo "  deps      - Download dependencies"
	@echo "  lint      - Run linter"
	@echo "  all       - deps + build"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  INSTALL_DIR=$(INSTALL_DIR)"