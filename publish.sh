#!/bin/bash

set -ex

KUBEVIRTCI_TAG=$(date +"%y%m%d%H%M")-$(git rev-parse --short HEAD)
export KUBEVIRTCI_TAG

TARGET_REPO="quay.io/kubevirtci"
TARGET_GIT_REMOTE="https://kubevirt-bot@github.com/kubevirt/kubevirtci.git"

# Build gocli
(cd cluster-provision/gocli && make container)
docker tag ${TARGET_REPO}/gocli ${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}

# Provision all base images
(cd cluster-provision/centos8 && ./build.sh)

# Provision all clusters
CLUSTERS="$(find cluster-provision/k8s/* -maxdepth 0 -type d -printf '%f\n')"
for i in ${CLUSTERS}; do
    cluster-provision/gocli/build/cli provision cluster-provision/k8s/$i
    docker tag ${TARGET_REPO}/k8s-$i ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}
done

# Push all images
IMAGES="${CLUSTERS}"
docker push ${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}
for i in ${IMAGES}; do
    TARGET_IMAGE="${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}"
    # until "unknown blob" issue is fixed use skopeo to push image
    skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
    #    docker push ${TARGET_IMAGE}
done

git config user.name "kubevirt-bot"
git config user.email "rmohr+kubebot@redhat.com"
git tag ${KUBEVIRTCI_TAG}
git push ${TARGET_GIT_REMOTE} ${KUBEVIRTCI_TAG}
