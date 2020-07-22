SHELL := /bin/bash


K8S118SUFFIX ?= :latest
K8S117SUFFIX ?= :latest
K8S116SUFFIX ?= :latest
K8S115SUFFIX ?= :latest
K8S114SIFFUX ?= :latest


BIN_DIR = $(CURDIR)/build
GO ?= go

export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off
export GOFLAGS=-mod=vendor


all: container-run

.PHONY: gocli
cli: fmt
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build \
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
	$(GO) fmt ./cmd/...
	$(GO) fmt ./docker/...

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
