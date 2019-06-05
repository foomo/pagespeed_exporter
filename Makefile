PREFIX              ?= $(shell pwd)
BIN_DIR             ?= $(shell pwd)

GO           := GO15VENDOREXPERIMENT=1 go
pkgs          = $(shell $(GO) list ./... | grep -v /vendor/)


all: format build test style

.PHONY: style
style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

.PHONY: test
test:
	@echo ">> running tests"
	@$(GO) test -short $(pkgs)

.PHONY: format
format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

.PHONY: vet
vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

.PHONY: build
build:
	@echo ">> building binaries"
	@go mod download
	@CGO_ENABLED=0 $(GO) build -ldflags "-X main.Version=`git rev-parse --short HEAD`" -o pagespeed_exporter pagespeed_exporter.go


.PHONY: release
release: goreleaser
	rm -f pagespeed_exporter
	goreleaser --rm-dist

.PHONY: goreleaser
goreleaser:
	@go get github.com/goreleaser/goreleaser && go install github.com/goreleaser/goreleaser

.PHONY: all style format build test vet tarball docker promu