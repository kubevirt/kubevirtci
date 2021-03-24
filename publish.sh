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
CLUSTERS="$(find cluster-provision/k8s/* -maxdepth 0 -type d -printf '%f\n')"
for i in ${CLUSTERS}; do
    cluster-provision/gocli/build/cli provision cluster-provision/k8s/$i
    docker tag ${TARGET_REPO}/k8s-$i ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}
done

# Provision 1.20-cgroupsv2 cluster
CGV2_CLUSTER="1.20"
CGV2_SUFFIX="cgroupsv2"
CGV2_IMAGE="${CGV2_CLUSTER}-${CGV2_SUFFIX}"
cluster-provision/gocli/build/cli provision \
    --cgroupv2=true --container-suffix=${CGV2_SUFFIX} \
    cluster-provision/k8s/${CGV2_CLUSTER}
docker tag ${TARGET_REPO}/k8s-${CGV2_IMAGE} \
    ${TARGET_REPO}/k8s-${CGV2_IMAGE}:${KUBEVIRTCI_TAG}

# Push all images
IMAGES="${CLUSTERS} ${CGV2_IMAGE}"
docker push ${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}
for i in ${IMAGES}; do
    docker push ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}
done

git config user.name "kubevirt-bot"
git config user.email "rmohr+kubebot@redhat.com"
git tag ${KUBEVIRTCI_TAG}
git push ${TARGET_GIT_REMOTE} ${KUBEVIRTCI_TAG}
