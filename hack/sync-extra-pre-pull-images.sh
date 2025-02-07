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
# Copyright 2025 The KubeVirt Authors.

GOCLI_MANIFESTS_DIR="$(pwd)/cluster-provision/gocli/opts/"

# Required images not present in gocli manifests
STATIC_IMAGES="quay.io/kubevirtci/install-cni:1.15.0
quay.io/kubevirtci/operator:1.15.0
quay.io/kubevirtci/pilot:1.15.0
quay.io/kubevirtci/proxyv2:1.15.0
quay.io/calico/cni:v3.26.5
quay.io/calico/kube-controllers:v3.26.5
quay.io/calico/node:v3.26.5"

while IFS=$'\n' read -r dir; do
    EXTRA_PREPULL_FILE="$dir/extra-pre-pull-images"
    if [ -f "$EXTRA_PREPULL_FILE" ]; then
        echo "Syncing provider: $dir"
        # aaq images and the sig-storage CSI images are not required for kubevirt E2E testing
        $dir/fetch-images.sh "$GOCLI_MANIFESTS_DIR" | grep -v sig-storage | grep -v aaq > $EXTRA_PREPULL_FILE
        echo "$STATIC_IMAGES" >> $EXTRA_PREPULL_FILE
    fi
done < <(find "./cluster-provision/k8s/" -maxdepth 1 -type d)
