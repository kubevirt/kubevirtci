#!/usr/bin/env bash
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

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
provision_dir="$DIR/$1"
[ -d "${provision_dir}" ] || exit 1

# compile the list of images from the manifests inside the version folder together with the
# gocli ones
{
    $DIR/fetch-images.sh "${provision_dir}" &
    $DIR/fetch-images.sh "$DIR/../gocli/opts/"
} |
    LC_ALL=C sort -u >"${provision_dir}/pre-pull-images"

# remove the duplicates that are already inside the pre-pull-images from the
# extra-pre-pull-images
for line in $(
    rg -N -f "${provision_dir}/extra-pre-pull-images" \
        "${provision_dir}/pre-pull-images"
); do
    sed -i "\#${line}#d" "${provision_dir}/extra-pre-pull-images"
done
