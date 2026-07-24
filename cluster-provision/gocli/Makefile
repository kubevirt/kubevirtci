SHELL := /bin/bash

IMAGES_FILE ?= images.json
KUBEVIRTCI_IMAGE_REPO ?= quay.io/kubevirtci
GOARCH ?= $$(uname -m | grep -q s390x && echo s390x || echo amd64)
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
	CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} $(GO) build -ldflags "-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.SUFFIX=:$(KUBEVIRTCI_TAG)'" -o $(BIN_DIR)/cli ./cmd/cli

.PHONY: fmt
fmt:
	$(GO) fmt ./cmd/...
	$(GO) fmt ./docker/...

.PHONY: container
container:
	podman build -f build/Dockerfile \
		--build-arg KUBEVIRTCI_TAG=$(KUBEVIRTCI_TAG) \
		-t ${KUBEVIRTCI_IMAGE_REPO}/gocli .

.PHONY: container-run
container-run: container
	podman run --rm ${KUBEVIRTCI_IMAGE_REPO}/gocli

.PHONY: vendor
vendor:
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: push
push:
	buildah manifest rm ${KUBEVIRTCI_IMAGE_REPO}/gocli:${KUBEVIRTCI_TAG} 2>/dev/null || true
	buildah build \
		-f build/Dockerfile \
		--platform $(MULTIARCH_PLATFORMS) \
		--build-arg KUBEVIRTCI_TAG=$(KUBEVIRTCI_TAG) \
		--manifest ${KUBEVIRTCI_IMAGE_REPO}/gocli:${KUBEVIRTCI_TAG} \
		.
	buildah manifest push --all \
		${KUBEVIRTCI_IMAGE_REPO}/gocli:${KUBEVIRTCI_TAG} \
		docker://${KUBEVIRTCI_IMAGE_REPO}/gocli:${KUBEVIRTCI_TAG}
