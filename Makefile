PREFIX              ?= $(shell pwd)
BIN_DIR             ?= $(shell pwd)

all: format build test

.PHONY: test
test:
	@echo ">> running tests"
	@go test -short ./...

.PHONY: format
format:
	@echo ">> formatting code"
	@go fmt

.PHONY: build
build:
	@echo ">> building binaries"
	@CGO_ENABLED=0 go build -ldflags "-X main.Version=`git rev-parse --short HEAD`" -o pagespeed_exporter pagespeed_exporter.go


.PHONY: release
release: goreleaser
	rm -f pagespeed_exporter
	goreleaser --rm-dist

.PHONY: goreleaser
goreleaser:
	@go get github.com/goreleaser/goreleaser && go install github.com/goreleaser/goreleaser

docker:
	docker build -t foomo/pagespeed_exporter:latest .