SHELL := /bin/bash

IMAGES_FILE ?= images.json

export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off
export GOFLAGS=-mod=vendor

BIN_DIR = $(CURDIR)/build
GO ?= go

all: container-run

.PHONY: gocli
cli:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags "-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.SUFFIX=:devel'" -o $(BIN_DIR)/cli ./cmd/cli
.PHONY: fmt
fmt:
	$(GO) fmt ./cmd/...
	$(GO) fmt ./docker/...

.PHONY: container
container: cli
	docker build -t kubevirtci/gocli build/
	docker tag kubevirtci/gocli kubevirtci/gocli:devel

.PHONY: container-run
container-run: container
	docker run kubevirtci/gocli

.PHONY: vendor
vendor:
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: push
push: container
	docker push kubevirtci/gocli
