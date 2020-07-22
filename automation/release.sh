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


function build_and_publish_base_images() {
    #TODO: Discover what base images need to be build
    for base_image in $base_images; do
        pushd cluster-provision/$base_image
            ./build.sh
            ./publish.sh
        popd
    done
}

function provision_and_publish_providers() {
    #TODO: Discover what providers need to be build
    for k8s_provider in $k8s_providers; do
        pushd cluster-provision/k8s/$k8s_provider
            ../provision.sh
            ../publish.sh
        popd
    done
}

#build_and_publish_base_images

#provision_and_publish_providers

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
export GO111MODULE=on
export GOFLAGS=""
go get github.com/github-release/github-release@v0.8.1

# Create the release
tag=$(date +%s)
org=kubevirt

git tag $tag

# To pass user/password from automations, idea is taken from [1]
# [1] https://stackoverflow.com/questions/8536732/can-i-hold-git-credentials-in-environment-variables
git config credential.helper '!f() { sleep 1; echo "username=${GITHUB_USER}"; echo "password=${GITHUB_TOKEN}"; }; f'


git push https://github.com/$org/kubevirtci.git $tag
github-release release \
        -u $org \
        -r kubevirtci \
        --tag $tag \
        --name $tag \
        --description "Follow instructions at kubevirtci.tar.gz README"

# Upload tarball
github-release upload \
        -u $org \
        -r kubevirtci \
        --name kubevirtci.tar.gz \
	    --tag $tag\
		--file $ARTIFACTS/kubevirtci.tar.gz


