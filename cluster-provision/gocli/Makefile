SHELL := /bin/bash

IMAGES_FILE ?= images.json
KUBEVIRTCI_IMAGE_REPO ?= quay.io/kubevirtci
GOARCH ?= $$(uname -m | grep -q s390x && echo s390x || echo amd64)

# Container runtime detection
CRI_BIN ?= $(shell command -v podman 2>/dev/null || command -v docker 2>/dev/null || echo "")
ifeq ($(CRI_BIN),)
$(error No container runtime found. Please install either docker or podman)
endif

export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off
export GOFLAGS=-mod=vendor

BIN_DIR = $(CURDIR)/build
GO ?= go

all: test-gocli container-run

.PHONY: test
test:
	$(GO) test -v ./cmd/... -coverprofile cover.out

.PHONY: test-gocli
test-gocli:
	$(GO) test -v ./... -coverprofile cover.out

.PHONY: gocli
cli:
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} $(GO) build -ldflags "-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.SUFFIX=:$(KUBEVIRTCI_TAG)'" -o $(BIN_DIR)/cli ./cmd/cli
.PHONY: fmt
fmt:
	$(GO) fmt ./cmd/...
	$(GO) fmt ./docker/...

.PHONY: container
container: cli
	$(CRI_BIN) build -t ${KUBEVIRTCI_IMAGE_REPO}/gocli build/

.PHONY: container-run
container-run: container
	$(CRI_BIN) run --rm ${KUBEVIRTCI_IMAGE_REPO}/gocli

.PHONY: vendor
vendor:
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: push
push: container
	$(CRI_BIN) push ${KUBEVIRTCI_IMAGE_REPO}/gocli
