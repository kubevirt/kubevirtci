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
# Copyright the KubeVirt Authors.

set -euo pipefail

SCRIPT_PATH=$(dirname "$(realpath "$0")")
source "$SCRIPT_PATH/vgpu_profile.sh"

if [ -z "${KUBEVIRTCI_PATH:-}" ]; then
    KUBEVIRTCI_PATH="$(cd "$SCRIPT_PATH/../.." && pwd)/"
fi
if [ -z "${KUBEVIRT_PROVIDER:-}" ]; then
    export KUBEVIRT_PROVIDER="k8s-1.35"
fi
if [[ -z "${KUBEVIRTCI_TAG:-}" && -z "${KUBEVIRTCI_GOCLI_CONTAINER:-}" ]]; then
    KUBEVIRTCI_TAG=$(curl --fail --silent --show-error --location \
        https://storage.googleapis.com/kubevirt-prow/release/kubevirt/kubevirtci/latest)
    export KUBEVIRTCI_TAG
fi

ssh_to_node() {
    local node_name=$1
    shift
    "${KUBEVIRTCI_PATH}/ssh.sh" "$node_name" "$@"
}

kubectl get kubevirt kubevirt -n kubevirt >/dev/null 2>&1 || {
    echo "KubeVirt is not installed in the active cluster. Install KubeVirt first." >&2
    exit 1
}

nodes=$(kubectl get nodes -l vgpu_capable=true -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}') || {
    echo "Could not list vgpu_capable=true nodes." >&2
    exit 1
}
if [[ -z "$nodes" ]]; then
    echo "No vgpu_capable=true nodes found. Run setup_vgpu_mdev.sh first." >&2
    exit 1
fi

for node in $nodes; do
    if ! ssh_to_node "$node" "lsmod | grep -q '^fake_nvidia_vgpu ' && $(vgpu_profile_check_command)"; then
        echo "Could not verify $node: check node access and the $VGPU_MDEV_TYPE and $VGPU_OVERRIDE_MDEV_TYPE driver profiles." >&2
        exit 1
    fi
    echo "$node exposes $VGPU_MDEV_TYPE ($VGPU_MDEV_NAME, max $VGPU_MAX_INSTANCES) and $VGPU_OVERRIDE_MDEV_TYPE ($VGPU_OVERRIDE_MDEV_NAME, max $VGPU_OVERRIDE_MAX_INSTANCES)."
done

echo "Configuring KubeVirt to manage fake NVIDIA mediated devices..."
kubectl patch kubevirt kubevirt -n kubevirt --type=merge -p "$(cat <<EOF
{
  "spec": {
    "configuration": {
      "mediatedDevicesConfiguration": {
        "mediatedDeviceTypes": ["$VGPU_MDEV_TYPE"],
        "nodeMediatedDeviceTypes": null,
        "enabled": true
      },
      "permittedHostDevices": {
        "mediatedDevices": [
          {"mdevNameSelector": "$VGPU_MDEV_NAME", "resourceName": "$VGPU_RESOURCE_NAME"},
          {"mdevNameSelector": "$VGPU_OVERRIDE_MDEV_NAME", "resourceName": "$VGPU_OVERRIDE_RESOURCE_NAME"}
        ]
      }
    }
  }
}
EOF
)"

echo "KubeVirt CR configured. Waiting for KubeVirt to become Available..."
kubectl wait kubevirt/kubevirt -n kubevirt --for=condition=Available --timeout=10m

generation=$(kubectl get kubevirt kubevirt -n kubevirt -o jsonpath='{.metadata.generation}')
if [[ ! "$generation" =~ ^[0-9]+$ ]]; then
    echo "KubeVirt did not report a valid generation: $generation" >&2
    exit 1
fi
expected_instances=$VGPU_MAX_INSTANCES
resource_name=$VGPU_RESOURCE_NAME

deadline=$((SECONDS + 600))
observed_generation=''
while (( SECONDS < deadline )); do
    observed_generation=$(kubectl get kubevirt kubevirt -n kubevirt -o jsonpath='{.status.observedGeneration}')
    if [[ "$observed_generation" =~ ^[0-9]+$ ]] && (( observed_generation >= generation )); then
        break
    fi
    sleep 5
done
if [[ ! "$observed_generation" =~ ^[0-9]+$ ]] || (( observed_generation < generation )); then
    echo "KubeVirt did not observe generation $generation within 10 minutes." >&2
    exit 1
fi

for node in $nodes; do
    echo "Waiting for $expected_instances $VGPU_MDEV_TYPE mdevs and $resource_name on $node..."
    mdev_count=''
    allocatable=''
    while (( SECONDS < deadline )); do
        handlers=$(kubectl get pods -n kubevirt -l kubevirt.io=virt-handler \
            --field-selector "spec.nodeName=$node,status.phase=Running" \
            -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')
        handler=${handlers%%$'\n'*}
        mdev_count=''
        if [[ -n "$handler" ]]; then
            mdev_count=$(kubectl exec -n kubevirt "$handler" -- bash -c '
                count=0
                for dev in /sys/bus/mdev/devices/*; do
                    [[ -e "$dev" ]] || continue
                    if [[ $(basename "$(readlink -f "$dev/mdev_type")") == "$1" ]]; then
                        count=$((count + 1))
                    fi
                done
                echo "$count"
            ' _ "$VGPU_MDEV_TYPE" 2>/dev/null) || mdev_count=''
        fi
        allocatable=$(kubectl get node "$node" \
            -o go-template="{{with index .status.allocatable \"$VGPU_RESOURCE_NAME\"}}{{.}}{{end}}")
        if [[ "$mdev_count" == "$expected_instances" && "$allocatable" == "$expected_instances" ]]; then
            echo "$node is ready: $mdev_count mdevs, $allocatable allocatable $resource_name devices."
            break
        fi
        sleep 5
    done
    if [[ "$mdev_count" != "$expected_instances" || "$allocatable" != "$expected_instances" ]]; then
        echo "$node did not become ready within 10 minutes (mdevs: ${mdev_count:-unknown}, allocatable: ${allocatable:-unknown})." >&2
        exit 1
    fi
done

echo "KubeVirt mediated devices are ready on all vGPU workers."
