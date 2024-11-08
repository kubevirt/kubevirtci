SHELL := /bin/bash

IMAGES_FILE ?= images.json
KUBEVIRTCI_IMAGE_REPO ?= quay.io/kubevirtci
GOARCH ?= $$(uname -m | grep -q s390x && echo s390x || echo amd64)

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
