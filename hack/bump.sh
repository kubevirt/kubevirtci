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
    if [ "$PROVIDER" == "gocli" ]; then
        sed -i "s#gocli@sha256:[a-z0-9]*#gocli@sha256:$HASH#" cluster-up/cluster/ephemeral-provider-common.sh
    else
        jq ".\"$PROVIDER\" = \"$HASH\"" cluster-provision/gocli/images.json > /tmp/images.json
        mv /tmp/images.json cluster-provision/gocli/images.json
    fi
}

main "$@"
