SHELL := /bin/bash

BIN_DIR = $(CURDIR)/build
GO ?= go

export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off
export GOFLAGS=-mod=vendor


all: push

.PHONY: fmt
fmt:
	$(GO) fmt ./check-image-pull-policies/...

.PHONY: container
container: fmt
	docker build -f check-image-pull-policies/Dockerfile -t kubevirtci/check-image-pull-policies .

.PHONY: vendor
vendor:
	$(GO) mod tidy
	$(GO) mod vendor

.PHONY: push
push: container
	docker push kubevirtci/check-image-pull-policies
