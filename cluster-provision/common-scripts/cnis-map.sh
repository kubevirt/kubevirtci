#!/usr/bin/env bash

set -e

declare -A CNI_MANIFESTS
CNI_MANIFESTS[1.17.0]="calico"
CNI_MANIFESTS[1.16.2]="flannel-ge-16.yaml"
CNI_MANIFESTS[1.15.1]="flannel-ge-16.yaml"
CNI_MANIFESTS[1.14.6]="flannel-ge-16.yaml"
CNI_MANIFESTS[1.13.3]="flannel-ge-12.yaml"
CNI_MANIFESTS[1.12.0]="flannel-ge-12.yaml"
CNI_MANIFESTS[1.11.0]="flannel.yaml"
CNI_MANIFESTS[1.10.11]="flannel.yaml"
export CNI_MANIFESTS
