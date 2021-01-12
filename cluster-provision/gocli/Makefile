SHELL := /bin/bash

IMAGES_FILE ?= images.json
KUBEVIRTCI_IMAGE_REPO ?= quay.io/kubevirtci

export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off
export GOFLAGS=-mod=vendor

BIN_DIR = $(CURDIR)/build
GO ?= go

all: container-run

.PHONY: gocli
cli:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags "-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.SUFFIX=:$(KUBEVIRTCI_TAG)'" -o $(BIN_DIR)/cli ./cmd/cli
.PHONY: fmt
fmt:
	$(GO) fmt ./cmd/...
	$(GO) fmt ./docker/...

.PHONY: container
container: cli
	docker build -t ${KUBEVIRTCI_IMAGE_REPO}/gocli build/

.PHONY: container-run
container-run: container
	docker run ${KUBEVIRTCI_IMAGE_REPO}/gocli

.PHONY: vendor
vendor:
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: push
push: container
	docker push ${KUBEVIRTCI_IMAGE_REPO}/gocli
