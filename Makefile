BIN_NAME ?= kvtools
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
PKG := github.com/coulof/kvtools/cmd

LDFLAGS := -s -w \
	-X $(PKG).Version=$(VERSION) \
	-X $(PKG).GitCommit=$(GIT_COMMIT) \
	-X $(PKG).BuildDate=$(BUILD_DATE)

.PHONY: all build build-all test vet clean install

all: build test

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BIN_NAME) main.go
	@echo "✔ Built bin/$(BIN_NAME) ($(VERSION))"

build-all:
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BIN_NAME)-linux-amd64 main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BIN_NAME)-linux-arm64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BIN_NAME)-darwin-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BIN_NAME)-darwin-arm64 main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BIN_NAME)-windows-amd64.exe main.go
	@echo "✔ Built multi-platform binaries in bin/ ($(VERSION))"

test:
	go test -v -race ./...

vet:
	go vet ./...

clean:
	rm -rf bin/

install: build
	install -m 755 bin/$(BIN_NAME) $(GOPATH)/bin/$(BIN_NAME)
	@echo "✔ Installed $(BIN_NAME) to $(GOPATH)/bin"
