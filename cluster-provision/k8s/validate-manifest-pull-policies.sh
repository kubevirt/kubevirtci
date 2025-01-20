#!/bin/bash

set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function usage {
  cat <<EOF
Usage: $0 <k8s-cluster-dir>

    Checks the input that is expected to be a deployment directory containing a directory called 'manifests'.

    Exits with non-zero exit code if the check fails

EOF
}

function detect_container_runtime() {
  if command -v podman &>/dev/null; then
    echo "podman"
  elif command -v docker &>/dev/null; then
    echo "docker"
  else
    echo "Error: Neither podman nor docker is available." >&2
    exit 1
  fi
}

function main {

  if [ "$#" -lt 1 ]; then
    usage
    exit 1
  fi

  manifest_dir="$DIR/$1/manifests"
  container_runtime=$(detect_container_runtime)

  echo "Checking $manifest_dir"
  $container_runtime run --rm -v "$manifest_dir:/manifests:Z" \
    quay.io/kubevirtci/check-image-pull-policies@sha256:c942d3a4a17f1576f81eba0a5844c904d496890677c6943380b543bbf2d9d1be \
    --manifest-source=/manifests \
    --dry-run=false \
    --verbose=false
}

main "$@"
