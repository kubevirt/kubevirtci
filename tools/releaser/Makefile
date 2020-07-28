SHELL := /bin/bash


BIN_DIR = $(CURDIR)/build
GO ?= go
GOARGS = GO111MODULE=on GOARCH=amd64 GOOS=linux CGO_ENABLED=0 GOFLAGS=-mod=vendor

all: releaaser

.PHONY: releaser
releaser:
	$(GOARGS) $(GO) build -o $(BIN_DIR)/releaser ./cmd/...

.PHONY: test
test :
	$(GOARGS) $(GO) test ./cmd/... -test.v -ginkgo.v

.PHONY: run
run: releaser
	$(BIN_DIR)/releaser $(WITH)

.PHONY: fmt
fmt:
	$(GOARGS) $(GO) fmt ./cmd/...

.PHONY: vendor
vendor:
	$(GO) mod tidy
	$(GO) mod vendor
