#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -exuo pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ksh="$(cd "$DIR/../.." && pwd)/cluster-up/kubectl.sh"
provision_dir="$1"
export KUBEVIRT_PROVIDER="k8s-${provision_dir}"

pre_pull_image_file="$DIR/${provision_dir}/pre-pull-images"
if [ ! -f "${pre_pull_image_file}" ]; then
    exit 1
fi

# deploy all manifests (except what is done on cluster-up or implictly already)
find "$DIR/${provision_dir}/manifests/" -type f -name '*.yaml' \
    -not -path '**/cnao/**' \
    -not -name 'logging.yaml' \
    -not -name 'local-volume.yaml' \
    -not -name 'cni.yaml' \
    -print -exec ${ksh} create -f {} \;

# wait for pods to get ready (we do this repeatedly to give the pods created by the operators time to come up)
repetition=0
until [ ${repetition} -ge 3 ]; do
    ${ksh} wait --for=condition=Ready pod --timeout=60s --all --all-namespaces
    sleep 10
    repetition=$((${repetition} + 1))
done
