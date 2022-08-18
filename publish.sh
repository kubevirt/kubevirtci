#!/bin/bash

set -ex

KUBEVIRTCI_TAG=$(date +"%y%m%d%H%M")-$(git rev-parse --short HEAD)
export KUBEVIRTCI_TAG

TARGET_REPO="quay.io/kubevirtci"
TARGET_KUBEVIRT_REPO="quay.io/kubevirt"
TARGET_GIT_REMOTE="https://kubevirt-bot@github.com/kubevirt/kubevirtci.git"

# Build gocli
(cd cluster-provision/gocli && make container)
docker tag ${TARGET_REPO}/gocli ${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}

# Provision all base images
(cd cluster-provision/centos8 && ./build.sh)

# Provision all clusters
CLUSTERS="$(find cluster-provision/k8s/* -maxdepth 0 -type d -printf '%f\n')"
for i in ${CLUSTERS}; do
    NETWORK_STACK="dualstack"
    if [[ $i =~ ipv6 ]]; then
        NETWORK_STACK="ipv6"
    fi

    cluster-provision/gocli/build/cli provision cluster-provision/k8s/$i --network-stack ${NETWORK_STACK}
    docker tag ${TARGET_REPO}/k8s-$i ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}

    cluster-provision/gocli/build/cli provision cluster-provision/k8s/$i --network-stack ${NETWORK_STACK} --slim
    docker tag ${TARGET_REPO}/k8s-$i ${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}-slim
done

# Provision alpine container disk and tag it
(cd cluster-provision/images/vm-image-builder && ./create-containerdisk.sh alpine-cloud-init)
docker tag  alpine-cloud-init:devel ${TARGET_REPO}/alpine-with-test-tooling-container-disk:${KUBEVIRTCI_TAG}
docker tag  alpine-cloud-init:devel ${TARGET_KUBEVIRT_REPO}/alpine-with-test-tooling-container-disk:devel

# Push all images

# until "unknown blob" issue is fixed use skopeo to push image
# see https://github.com/moby/moby/issues/43234

IMAGES="${CLUSTERS}"
TARGET_IMAGE="${TARGET_REPO}/gocli:${KUBEVIRTCI_TAG}"
skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
#docker push ${TARGET_IMAGE}
for i in ${IMAGES}; do
    TARGET_IMAGE="${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}"
    skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
    #docker push ${TARGET_IMAGE}

    TARGET_IMAGE="${TARGET_REPO}/k8s-$i:${KUBEVIRTCI_TAG}-slim"
    skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
done

TARGET_IMAGE="${TARGET_REPO}/alpine-with-test-tooling-container-disk:${KUBEVIRTCI_TAG}"
skopeo copy "docker-daemon:${TARGET_IMAGE}" "docker://${TARGET_IMAGE}"
#docker push ${TARGET_IMAGE}
TARGET_KUBEVIRT_IMAGE="${TARGET_KUBEVIRT_REPO}/alpine-with-test-tooling-container-disk:devel"
skopeo copy "docker-daemon:${TARGET_KUBEVIRT_IMAGE}" "docker://${TARGET_KUBEVIRT_IMAGE}"

git config user.name "kubevirt-bot"
git config user.email "rmohr+kubebot@redhat.com"
git tag ${KUBEVIRTCI_TAG}
git push ${TARGET_GIT_REMOTE} ${KUBEVIRTCI_TAG}
