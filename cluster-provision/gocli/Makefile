SHELL := /bin/bash

BIN_DIR = $(CURDIR)/build

export GOPROXY=direct
export GOSUMDB=off
export GOFLAGS=-mod=vendor
export GOROOT=$(BIN_DIR)/go/
export GOBIN=$(GOROOT)/bin/
export PATH := $(GOROOT)/bin:$(PATH)
export GO := $(GOBIN)/go

all: fmt build

$(GO):
	hack/install-go.sh $(BIN_DIR)

.PHONY: gocli
cli: fmt
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o $(BIN_DIR)/cli ./cmd/cli

.PHONY: fmt
fmt: $(GO)
	$(GO) fmt ./cmd/...
	$(GO) fmt ./docker/...

.PHONY: container
container: cli
	docker build -t kubevirtci/gocli build/

.PHONY: container-run
container-run: container
	docker run kubevirtci/gocli

.PHONY: vendor
vendor: $(GO)
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: push
push: container
	docker push kubevirtci/gocli
