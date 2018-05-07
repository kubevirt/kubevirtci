all: fmt build

build: fmt
	bazel build //gocli:cli

fmt:
	go fmt gocli/cmd/...
	go fmt gocli/docker/...

container: fmt
	bazel build //gocli:gocli

container-run: fmt
	bazel run //gocli:gocli -- ${ARGS}

push: fmt
	bazel run //:push-all

generate:
	dep ensure
	bazel run //:gazelle
