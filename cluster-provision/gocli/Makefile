SHELL := /bin/bash


K8S118SUFFIX ?= :latest
K8S117SUFFIX ?= :latest
K8S116SUFFIX ?= :latest
K8S115SUFFIX ?= :latest
K8S114SUFFIX ?= :latest


BIN_DIR = $(CURDIR)/build
GO ?= go

GOARGS=GOFLAGS=-mod=vendor CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on GOPROXY=direct GOSUMDB=off

all: container-run

.PHONY: gocli
cli:
	$(GOARGS) $(GO) build \
				-ldflags "\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S118SUFFIX=$(K8S118SUFFIX)'\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S117SUFFIX=$(K8S117SUFFIX)'\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S116SUFFIX=$(K8S116SUFFIX)'\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S115SUFFIX=$(K8S115SUFFIX)'\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S114SUFFIX=$(K8S114SUFFIX)'\
				" \
				-o $(BIN_DIR)/cli ./cmd/cli

.PHONY: fmt
fmt:
	$(GOARGS) $(GO) fmt ./cmd/... ./docker/...

.PHONY: container
container: cli
	docker build -t kubevirtci/gocli build/

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
