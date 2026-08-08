SHELL := /bin/bash

IMAGES_FILE ?= images.json
KUBEVIRTCI_IMAGE_REPO ?= quay.io/kubevirtci
GOCLI_IMAGE ?= $(KUBEVIRTCI_IMAGE_REPO)/gocli
GOARCH ?= $$(uname -m | grep -q s390x && echo s390x || echo amd64)
MULTIARCH_PLATFORMS ?= linux/amd64,linux/s390x

export GO111MODULE=on
export GOPROXY=direct
export GOFLAGS=-mod=vendor

BIN_DIR = $(CURDIR)/build
CRI_BIN ?= podman
GO      ?= go

all: test-gocli container-run

.PHONY: test
test:
	$(GO) test -v ./cmd/... -coverprofile cover.out

.PHONY: test-gocli
test-gocli:
	$(GO) test -v ./... -coverprofile cover.out

.PHONY: cli
cli:
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} $(GO) build -ldflags "-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.SUFFIX=:$(KUBEVIRTCI_TAG)'" -o $(BIN_DIR)/cli-${GOARCH} ./cmd/cli
	ln -sf cli-${GOARCH} $(BIN_DIR)/cli

.PHONY: fmt
fmt:
	$(GO) fmt ./cmd/...
	$(GO) fmt ./docker/...

.PHONY: container
container: cli
	$(CRI_BIN) image build --build-arg TARGETARCH=${GOARCH} -t $(GOCLI_IMAGE) build/

.PHONY: container-run
container-run: container
	$(CRI_BIN) container run --rm $(GOCLI_IMAGE)

.PHONY: vendor
vendor:
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: push
push:
	$(MAKE) GOARCH=s390x cli
	$(MAKE) GOARCH=amd64 cli
	$(CRI_BIN) image rm $(GOCLI_IMAGE):$(KUBEVIRTCI_TAG) 2>/dev/null || true
	$(CRI_BIN) manifest rm $(GOCLI_IMAGE):$(KUBEVIRTCI_TAG) 2>/dev/null || true
	$(CRI_BIN) manifest create $(GOCLI_IMAGE):$(KUBEVIRTCI_TAG)
	$(CRI_BIN) image build ./build/ \
		--platform $(MULTIARCH_PLATFORMS) \
		--manifest $(GOCLI_IMAGE):$(KUBEVIRTCI_TAG)
	$(CRI_BIN) manifest push $(GOCLI_IMAGE):$(KUBEVIRTCI_TAG) docker://$(GOCLI_IMAGE):$(KUBEVIRTCI_TAG)
