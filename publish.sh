#!/bin/bash

set -ex

export KUBEVIRTCI_TAG=$(date +"%y%m%d%H%M")-$(git rev-parse --short HEAD)

TARGET_REPO="quay.io/kubevirtci"
TARGET_KUBEVIRT_REPO="quay.io/kubevirt"
TARGET_GIT_REMOTE="https://kubevirt-bot@github.com/kubevirt/kubevirtci.git"

function build_gocli() {
  (cd cluster-provision/gocli && make container)
  docker tag ${TARGET_REPO}/gocli ${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}
}

function build_centos8_base_image() {
  (cd cluster-provision/centos8 && ./build.sh)
}

function build_centos9_base_image() {
  (cd cluster-provision/centos9 && ./build.sh)
}

function build_base_images() {
  build_centos8_base_image
  build_centos9_base_image
}

function build_clusters() {
  CLUSTERS="$(find cluster-provision/k8s/* -maxdepth 0 -type d -printf '%f\n')"
  for i in ${CLUSTERS}; do
    cluster-provision/gocli/build/cli provision cluster-provision/k8s/$i
    docker tag ${TARGET_REPO}/k8s-$i ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}

    cluster-provision/gocli/build/cli provision cluster-provision/k8s/$i --slim
    docker tag ${TARGET_REPO}/k8s-$i ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}-slim
  done
}

function push_cluster_images() {
  # until "unknown blob" issue is fixed use skopeo to push image
  # see https://github.com/moby/moby/issues/43234
  IMAGES="${CLUSTERS}"
  for i in ${IMAGES}; do
    TARGET_IMAGE="${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}"
    skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"

    TARGET_IMAGE="${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}-slim"
    skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
  done
}

function push_gocli() {
  # push gocli image
  TARGET_IMAGE="${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}"
  skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
}

function publish_clusters() {
  build_clusters
  push_cluster_images
}

function build_alpine_container_disk() {
  (cd cluster-provision/images/vm-image-builder && ./create-containerdisk.sh alpine-cloud-init)
  docker tag alpine-cloud-init:devel ${TARGET_REPO}/alpine-with-test-tooling-container-disk:${KUBEVIRTCI_TAG}
  docker tag alpine-cloud-init:devel ${TARGET_KUBEVIRT_REPO}/alpine-with-test-tooling-container-disk:devel
}

function push_alpine_container_disk() {
  TARGET_IMAGE="${TARGET_REPO}/alpine-with-test-tooling-container-disk:${KUBEVIRTCI_TAG}"
  skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
  TARGET_KUBEVIRT_IMAGE="${TARGET_KUBEVIRT_REPO}/alpine-with-test-tooling-container-disk:devel"
  skopeo copy "docker-daemon:${TARGET_KUBEVIRT_IMAGE}" "docker://${TARGET_KUBEVIRT_IMAGE}"
}

function publish_alpine_container_disk() {
  build_alpine_container_disk
  push_alpine_container_disk
}

function create_git_tag() {
  git config user.name "kubevirt-bot"
  git config user.email "rmohr+kubebot@redhat.com"
  git tag ${KUBEVIRTCI_TAG}
  git push ${TARGET_GIT_REMOTE} ${KUBEVIRTCI_TAG}
}

function main() {
  build_gocli
  build_base_images
  publish_clusters
  publish_alpine_container_disk
  push_gocli
  create_git_tag
}

main "$@"
