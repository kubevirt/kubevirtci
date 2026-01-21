#!/bin/bash

set -e

function detect_cri() {
    if podman ps >/dev/null 2>&1; then echo podman; elif docker ps >/dev/null 2>&1; then echo docker; fi
}

export KUBEVIRTCI_TAG=${KUBEVIRTCI_TAG:-$(date +"%y%m%d%H%M")-$(git rev-parse --short HEAD)}
export CRI_BIN=${CRI_BIN:-$(detect_cri)}

if [ -z "${CRI_BIN}" ]; then
    echo "ERROR: Neither podman nor docker is available." >&2
    exit 1
fi

TARGET_REPO="quay.io/kubevirtci"
TARGET_KUBEVIRT_REPO="quay.io/kubevirt"

function build_alpine_container_disk() {
  echo "INFO: build alpine container disk"
  (cd cluster-provision/images/vm-image-builder && ./create-containerdisk.sh alpine-cloud-init)
  ${CRI_BIN} tag alpine-cloud-init:devel ${TARGET_REPO}/alpine-with-test-tooling-container-disk:${KUBEVIRTCI_TAG}
  ${CRI_BIN} tag alpine-cloud-init:devel ${TARGET_KUBEVIRT_REPO}/alpine-with-test-tooling-container-disk:devel
}

function push_alpine_container_disk() {
  echo "INFO: push alpine container disk"
  TARGET_IMAGE="${TARGET_REPO}/alpine-with-test-tooling-container-disk:${KUBEVIRTCI_TAG}"
  ${CRI_BIN} push $TARGET_IMAGE
  TARGET_KUBEVIRT_IMAGE="${TARGET_KUBEVIRT_REPO}/alpine-with-test-tooling-container-disk:devel"
  ${CRI_BIN} push $TARGET_KUBEVIRT_IMAGE
}

function publish_alpine_container_disk() {
  build_alpine_container_disk
  push_alpine_container_disk
}

publish_alpine_container_disk
