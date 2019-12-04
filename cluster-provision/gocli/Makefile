all: fmt build

.PHONY: build
build: fmt
	bazel run //:gazelle
	bazel build //:cli

.PHONY: fmt
fmt:
	go fmt ./cmd/...
	go fmt ./docker/...

.PHONY: container
container: fmt
	bazel build //:gocli

.PHONY: container-run
container-run: fmt
	bazel run //:gocli

.PHONY: deps-update
deps-update:
	GO111MODULE=on go mod tidy
	GO111MODULE=on go mod vendor
	bazel run //:gazelle

.PHONY: push
push: fmt
	bazel run --host_force_python=PY2 //:push-all
