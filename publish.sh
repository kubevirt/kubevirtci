#!/bin/bash

set -ex

export KUBEVIRTCI_TAG=$(date +"%y%m%d%H%M")-$(git rev-parse --short HEAD)

TARGET_REPO="quay.io/kubevirtci"
TARGET_GIT_REMOTE="https://kubevirt-bot@github.com/kubevirt/kubevirtci.git"

# Build gocli
(cd cluster-provision/gocli && make container)
docker tag ${TARGET_REPO}/gocli ${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}

# Provision all base images
(cd cluster-provision/centos8 && ./build.sh)

# Provision all clusters
for i in $(find cluster-provision/k8s/* -maxdepth 0 -type d -printf '%f\n'); do
    cluster-provision/gocli/build/cli provision cluster-provision/k8s/$i
    docker tag ${TARGET_REPO}/k8s-$i ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}
done

# Push all images
docker push ${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}
for i in $(find cluster-provision/k8s/* -maxdepth 0 -type d -printf '%f\n'); do
    docker push ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}
done

git config user.name "kubevirt-bot"
git config user.email "rmohr+kubebot@redhat.com"
git tag ${KUBEVIRTCI_TAG}
git push ${TARGET_GIT_REMOTE} ${KUBEVIRTCI_TAG}
