BIN_NAME ?= kvtools
PLUGIN_NAME ?= kubectl-kvtools
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
PKG := github.com/coulof/kvtools/cmd

LDFLAGS := -s -w \
	-X $(PKG).Version=$(VERSION) \
	-X $(PKG).GitCommit=$(GIT_COMMIT) \
	-X $(PKG).BuildDate=$(BUILD_DATE)

.PHONY: all build test vet clean install docker-build

all: build test

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BIN_NAME) main.go
	@cp bin/$(BIN_NAME) bin/$(PLUGIN_NAME)
	@echo "✔ Built bin/$(BIN_NAME) and bin/$(PLUGIN_NAME) ($(VERSION))"

test:
	go test -v -race ./...

vet:
	go vet ./...

clean:
	rm -rf bin/

install: build
	install -m 755 bin/$(BIN_NAME) $(GOPATH)/bin/$(BIN_NAME)
	install -m 755 bin/$(PLUGIN_NAME) $(GOPATH)/bin/$(PLUGIN_NAME)
	@echo "✔ Installed $(BIN_NAME) and $(PLUGIN_NAME) to $(GOPATH)/bin"

docker-build:
	docker build -t coulof/kvtools:$(VERSION) -t coulof/kvtools:latest .
