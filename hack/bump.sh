#!/bin/bash -e
#
# This file is part of the KubeVirt project
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
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
# Copyright 2020 Red Hat, Inc.
#

# Updates cluster-up/cluster/images.sh provider hash to point on new given hash.
# After usage, commit the changes of cluster-up/cluster/images.sh.

# Usage: ./hack/bump.sh <provider> <hash>
# Example: ./hack/bump.sh k8s-1.18 c41e3d9adb756b60e1fbce2ffd774c66c99fdde7ee337460c472ee92868e579e

PROVIDER=${1:?}
HASH=${2:?}

function main() {
    sed -i "s/IMAGES\["$PROVIDER"\].*/IMAGES\["$PROVIDER"\]=\""$PROVIDER"@sha256:"$HASH"\"/g" cluster-up/cluster/images.sh
}

main "$@"
