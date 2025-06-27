#!/bin/bash
#
# This file is part of the KubeVirt project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Copyright The KubeVirt Authors.
#
#

set -exuo pipefail

auto_update=
if [ $# -gt 1 ] && [ "$1" == "--auto-update" ]; then
    auto_update=true
    shift
fi

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ksh="$(cd "$DIR/../.." && pwd)/cluster-up/kubectl.sh"
provision_dir="$1"
export KUBEVIRT_PROVIDER="k8s-${provision_dir}"

pre_pull_image_file="$DIR/${provision_dir}/pre-pull-images"
if [ ! -f "${pre_pull_image_file}" ]; then
    echo "$DIR/${provision_dir}/pre-pull-images not found!"
    exit 1
fi

extra_pre_pull_image_file="$DIR/${provision_dir}/extra-pre-pull-images"
if [ ! -f "${extra_pre_pull_image_file}" ]; then
    echo "$DIR/${provision_dir}/extra-pre-pull-images not found!"
    exit 1
fi

# check image version for pods
images_not_in_list=$(mktemp)
trap 'rm -f $images_not_in_list' EXIT SIGINT SIGTERM
for image in $(${ksh} get pods --all-namespaces -o jsonpath="{..image}" | tr -s '[[:space:]]' '\n' | grep -v 'registry.k8s.io' | sort | uniq); do
    set +e
    if ! grep -q "$image" "${pre_pull_image_file}" && ! grep -q "$image" "${extra_pre_pull_image_file}"; then
        echo "$image" >>"${images_not_in_list}"
    fi
    set -e
done
if [ -s "${images_not_in_list}" ]; then
    echo "Images found in cluster that are not in list!"
    if [[ "$auto_update" != "true" ]]; then
        cat "${images_not_in_list}"
        echo "(Please add them to file ${extra_pre_pull_image_file})"
        exit 1
    else
        cat "${images_not_in_list}" >>"${extra_pre_pull_image_file}"
        echo "${extra_pre_pull_image_file} updated with images not in list."
    fi
else
    echo "No images found in cluster that are not in list."
fi
