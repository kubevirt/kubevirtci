#!/bin/bash

set -e

export KUBEVIRTCI_TAG=$(date +"%y%m%d%H%M")-$(git rev-parse --short HEAD)

TARGET_REPO="quay.io/kubevirtci"
TARGET_KUBEVIRT_REPO="quay.io/kubevirt"
TARGET_GIT_REMOTE="https://kubevirt-bot@github.com/kubevirt/kubevirtci.git"

function detect_cri() {
    if podman ps >/dev/null 2>&1; then echo podman; elif docker ps >/dev/null 2>&1; then echo docker; fi
}

export CRI_BIN=${CRI_BIN:-$(detect_cri)}
export DRY_RUN=${DRY_RUN:-false}

PREV_KUBEVIRTCI_TAG=$(curl -sL https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirtci/latest?ignoreCache=1)
CLUSTERS="$(find cluster-provision/k8s/* -maxdepth 0 -type d -printf '%f\n')"
IMAGES_TO_BUILD=()
IMAGES_TO_RETAG=()

function run_provision_manager() {
  # Run provision manager to determine which images to build
  TARGETS=$(./hack/pman/pman.sh)
  echo "INFO: Provision manager suggests to build the following providers: ${TARGETS[@]}"

  declare -A targets_map
  for target in ${TARGETS[@]}; do
    targets_map[$target]=$target
  done

  for i in ${CLUSTERS}; do
    if [ "${targets_map[$i]}" != "" ]; then
      IMAGES_TO_BUILD+=($i)
      echo "INFO: $i will be provisioned"
    else
      IMAGES_TO_RETAG+=($i)
      echo "INFO: $i will be retagged ($PREV_KUBEVIRTCI_TAG)"
    fi
  done

  if [ $DRY_RUN == true ]; then
    exit 0
  fi
}

function build_gocli() {
  (cd cluster-provision/gocli && make container)
  ${CRI_BIN} tag ${TARGET_REPO}/gocli ${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}
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
  for i in ${IMAGES_TO_BUILD[@]}; do
    echo "INFO: building $i"
    cluster-provision/gocli/build/cli provision cluster-provision/k8s/$i
    ${CRI_BIN} tag ${TARGET_REPO}/k8s-$i ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}

    cluster-provision/gocli/build/cli provision cluster-provision/k8s/$i --slim
    ${CRI_BIN} tag ${TARGET_REPO}/k8s-$i ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}-slim
  done
}

function push_cluster_images() {
  # until "unknown blob" issue is fixed use skopeo to push image
  # see https://github.com/moby/moby/issues/43234
  for i in ${IMAGES_TO_BUILD[@]}; do
    echo "INFO: push $i"
    TARGET_IMAGE="${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}"
    skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"

    TARGET_IMAGE="${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}-slim"
    skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
  done

  # images that the change doesn't affect can be retagged from previous tag
  for i in ${IMAGES_TO_RETAG[@]}; do
    echo "INFO: retagging $i (previous tag $PREV_KUBEVIRTCI_TAG)"
    skopeo copy "docker://${TARGET_REPO}/k8s-$i:${PREV_KUBEVIRTCI_TAG}" "docker://${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}"
    skopeo copy "docker://${TARGET_REPO}/k8s-$i:${PREV_KUBEVIRTCI_TAG}-slim" "docker://${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}-slim"
  done
}

function push_gocli() {
  echo "INFO: push gocli"
  TARGET_IMAGE="${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}"
  skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
}

function publish_clusters() {
  build_clusters
  push_cluster_images
}

function build_alpine_container_disk() {
  echo "INFO: build alpine container disk"
  (cd cluster-provision/images/vm-image-builder && ./create-containerdisk.sh alpine-cloud-init)
  ${CRI_BIN} tag alpine-cloud-init:devel ${TARGET_REPO}/alpine-with-test-tooling-container-disk:${KUBEVIRTCI_TAG}
  ${CRI_BIN} tag alpine-cloud-init:devel ${TARGET_KUBEVIRT_REPO}/alpine-with-test-tooling-container-disk:devel
}

function push_alpine_container_disk() {
  echo "INFO: push alpine container disk"
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
  echo "INFO: pushed new tag $KUBEVIRTCI_TAG"
}

function main() {
  run_provision_manager
  build_gocli
  build_base_images
  publish_clusters
  publish_alpine_container_disk
  push_gocli
  create_git_tag
}

main "$@"
