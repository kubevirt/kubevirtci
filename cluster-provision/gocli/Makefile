SHELL := /bin/bash


K8S118SHA ?= c41e3d9adb756b60e1fbce2ffd774c66c99fdde7ee337460c472ee92868e579e
K8S117SHA ?= 7506c49307944d2bbbf6486bc9c9aa4496694d3284143a51ee87333c1a0ab7f0
K8S116SHA ?= 1e153fb62c9a30ce6bc6ddc1af21bb28a56b780ec93ab15e113d729bf664469a
K8S115SHA ?= c58cb9d79968590f24e070bc2517088d44fa2f83ba73e989a7e0f690ad08460b
K8S114SHA ?= 46e449b292dcb420f0944cac0a7a5c667c6f19bba2a4192737380e8e77f27ed0


BIN_DIR = $(CURDIR)/build
GO ?= go

export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off
export GOFLAGS=-mod=vendor


all: container-run

.PHONY: gocli
cli: fmt
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build \
				-ldflags "\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S118SHA=$(K8S118SHA)'\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S117SHA=$(K8S117SHA)'\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S116SHA=$(K8S116SHA)'\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S115SHA=$(K8S115SHA)'\
					-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.K8S114SHA=$(K8S114SHA)'\
				" \
				-o $(BIN_DIR)/cli ./cmd/cli

.PHONY: fmt
fmt:
	@$(GO) fmt ./cmd/...
	@$(GO) fmt ./docker/...

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
