#!/bin/bash -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

centos_version="$(cat $DIR/version | tr -d '\n')"

container_runtime=""
if command -v podman >/dev/null 2>&1; then
  container_runtime="podman"
elif command -v docker >/dev/null 2>&1; then
  container_runtime="docker"
else
  echo "Error: No container runtime found. Install podman or docker first." >&2
  exit 1
fi

$container_runtime build --build-arg BUILDARCH=$(uname -m) --build-arg centos_version=$centos_version . -t quay.io/kubevirtci/centos9