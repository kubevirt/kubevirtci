#!/bin/bash

set -euxo pipefail

workdir=$(mktemp -d)
ARTIFACTS=${ARTIFACTS:-/tmp}

end() {
    rm -rf $workdir
}
trap end EXIT


function get_latest_digest_suffix() {
    local provider=$1
    local latest_digest=$(docker run alexeiled/skopeo skopeo inspect docker://docker.io/kubevirtci/$provider:latest | docker run -i imega/jq -r -c .Digest)
    echo "@$latest_digest"
}

#TODO: discover what providers has changed and re-provision there
#pushd cluster-provision/k8s/1.18
#    ../provision.sh
#    ..publish.sh
#popd

pushd cluster-provision/gocli
    make cli \
        K8S118SUFFIX="$(get_latest_digest_suffix k8s-1.18)" \
        K8S117SUFFIX="$(get_latest_digest_suffix k8s-1.17)" \
        K8S116SUFFIX="$(get_latest_digest_suffix k8s-1.16)" \
        K8S115SUFFIX="$(get_latest_digest_suffix k8s-1.15)" \
        K8S114SUFFIX="$(get_latest_digest_suffix k8s-1.14)"
popd

# Install cluster-up
cp -rf cluster-up/* $workdir

# Install gocli
cp -f cluster-provision/gocli/build/cli  $workdir

# Create the tarball
tar -C $workdir -cvzf $ARTIFACTS/kubevirtci.tar.gz .

# TODO release tarball
