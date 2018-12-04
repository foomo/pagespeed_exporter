# Copyright 2015 The Prometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO           := GO15VENDOREXPERIMENT=1 go
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
PROMU        := $(FIRST_GOPATH)/bin/promu
pkgs          = $(shell $(GO) list ./... | grep -v /vendor/)

PREFIX              ?= $(shell pwd)
BIN_DIR             ?= $(shell pwd)
DOCKER_IMAGE_NAME   ?= foomo/pagespeed_exporter
DOCKER_IMAGE_TAG    ?= latest

all: format build test

style:
	@echo ">> checking code style"
	@! gofmt -d $(shell find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

test:
	@echo ">> running tests"
	@$(GO) test -short -race $(pkgs)

format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

build: dep
	@echo ">> building binaries"
	@dep ensure
	@CGO_ENABLED=0 $(GO) build -ldflags "-X main.Version=`git rev-parse --short HEAD`" -o pagespeed_exporter pagespeed_exporter.go

dep:
ifeq ($(shell command -v dep 2> /dev/null),)
	go get -u -v github.com/golang/dep/cmd/dep
endif

docker-build:
	@echo ">> building docker image"
	@docker build -t "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

docker-push:
	@echo ">> pushing docker image"
	@docker login -u="$(DOCKER_USERNAME)" -p="$(DOCKER_PASSWORD)"
	@docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
	@docker push $(DOCKER_IMAGE_NAME):latest

.PHONY: all style format build test vet tarball docker promu