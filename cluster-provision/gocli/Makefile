SHELL := /bin/bash

IMAGES_FILE ?= images.json
KUBEVIRTCI_IMAGE_REPO ?= quay.io/kubevirtci
GOARCH ?= $$(uname -m | grep -q s390x && echo s390x || echo amd64)
BUILDX_BUILDER ?= gocli-multiarch
MULTIARCH_PLATFORMS ?= linux/amd64,linux/s390x

export GO111MODULE=on
export GOPROXY=direct
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

.PHONY: cli
cli:
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} $(GO) build -ldflags "-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.SUFFIX=:$(KUBEVIRTCI_TAG)'" -o $(BIN_DIR)/cli-${GOARCH} ./cmd/cli

.PHONY: fmt
fmt:
	$(GO) fmt ./cmd/...
	$(GO) fmt ./docker/...

.PHONY: container
container: cli
	docker build --build-arg TARGETARCH=${GOARCH} -t ${KUBEVIRTCI_IMAGE_REPO}/gocli build/

.PHONY: container-run
container-run: container
	docker run --rm ${KUBEVIRTCI_IMAGE_REPO}/gocli

.PHONY: vendor
vendor:
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: push
push:
	GOARCH=s390x $(MAKE) cli
	GOARCH=amd64 $(MAKE) cli
	# ensure a dedicated multi-platform builder exists
	docker buildx inspect $(BUILDX_BUILDER) >/dev/null 2>&1 || \
		docker buildx create --name $(BUILDX_BUILDER) --use
	docker buildx build --builder $(BUILDX_BUILDER) \
		--platform $(MULTIARCH_PLATFORMS) \
		--provenance=false \
		-t ${KUBEVIRTCI_IMAGE_REPO}/gocli:${KUBEVIRTCI_TAG} \
		--push build/
