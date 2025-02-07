#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -exuo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ksh="$(cd "$DIR/../.." && pwd)/cluster-up/kubectl.sh"
provision_dir="$1"
export KUBEVIRT_PROVIDER="k8s-${provision_dir}"

pre_pull_image_file="$DIR/${provision_dir}/extra-pre-pull-images"
if [ ! -f "${pre_pull_image_file}" ]; then
    exit 1
fi

# check image version for pods
images_not_in_list=$(mktemp)
images_from_manifests=$(mktemp)
trap 'rm -f $images_not_in_list $images_from_manifests' EXIT SIGINT SIGTERM
$DIR/fetch-images.sh "$DIR/${provision_dir}" > "${images_from_manifests}"
$DIR/fetch-images.sh "$DIR/../gocli/opts/" >> "${images_from_manifests}"
for image in $(${ksh} get pods --all-namespaces -o jsonpath="{..image}" | tr -s '[[:space:]]' '\n' | grep -v 'registry.k8s.io' | sort | uniq); do
    set +e
    if ! grep -q "$image" "${pre_pull_image_file}"; then
        if ! grep -q "$image" "${images_from_manifests}"; then
            echo "$image" >>"${images_not_in_list}"
        fi
    fi
    set -e
done
if [ -s "${images_not_in_list}" ]; then
    echo "Images found in cluster that are not in list!"
    cat "${images_not_in_list}"
    echo "(Please add them to file ${pre_pull_image_file})"
    exit 1
fi
