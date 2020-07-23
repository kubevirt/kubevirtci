#!/bin/bash

set -euxo pipefail

workdir=$(mktemp -d)
ARTIFACTS=${ARTIFACTS:-/tmp}
base_images=(centos8)
k8s_providers=(1.17 1.18)

end() {
    rm -rf $workdir
}
trap end EXIT


function get_latest_digest_suffix() {
    local provider=$1
    local latest_digest=$(docker run alexeiled/skopeo skopeo inspect docker://docker.io/kubevirtci/$provider:latest | docker run -i imega/jq -r -c .Digest)
    echo "@$latest_digest"
}


#TODO: Discover what base images need to be build
for base_image in $base_images; do
    pushd cluster-provision/$base_image
        ./build.sh
        ./publish.sh
    popd
done

#TODO: Discover what providers need to be build
for k8s_provider in $k8s_providers; do
    pushd cluster-provision/k8s/$k8s_provider
        ../provision.sh
        ../publish.sh
    popd
done

pushd cluster-provision/gocli
    make cli \
        K8S118SUFFIX="$(get_latest_digest_suffix k8s-1.18)" \
        K8S117SUFFIX="$(get_latest_digest_suffix k8s-1.17)" \
        K8S116SUFFIX="$(get_latest_digest_suffix k8s-1.16)" \
        K8S115SUFFIX="$(get_latest_digest_suffix k8s-1.15)" \
        K8S114SUFFIX="$(get_latest_digest_suffix k8s-1.14)"
popd

# Create kubevirtci dir inside the tarball
mkdir $workdir/kubevirtci

# Install cluster-up
cp -rf cluster-up/* $workdir/kubevirtci

# Install gocli
cp -f cluster-provision/gocli/build/cli  $workdir/kubevirtci

# Create the tarball
tar -C $workdir -cvzf $ARTIFACTS/kubevirtci.tar.gz .

# Install github-release tool
# TODO: Vendor this
go get github.com/github-release/github-release@v0.8.1

# Create the release
tag $(git rev-parse --short HEAD)
github-release release \
        -u kubevirt \
        -r kubevirtci \
        --tag $tag \
        --name $tag \
        --description "Follow instructions at kubevirtci.tar.gz README"

# Upload tarball
github-release upload \
        -u kubevirt \
        -r kubevirtci \
        --name kubevirtci.tar.gz \
	    --tag $tag\
		--file $ARTIFACTS/kubevirtci.tar.gz


