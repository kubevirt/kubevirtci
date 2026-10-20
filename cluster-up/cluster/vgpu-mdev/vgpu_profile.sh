#!/usr/bin/env bash

# The default profile KubeVirt creates. The E2E node override test uses 223.
VGPU_MDEV_TYPE=nvidia-222
VGPU_MDEV_NAME='GRID T4-1B'
VGPU_RESOURCE_NAME=nvidia.com/GRID_T4-1B
VGPU_MAX_INSTANCES=16
VGPU_OVERRIDE_MDEV_TYPE=nvidia-223
VGPU_OVERRIDE_MDEV_NAME='GRID T4-2B'
VGPU_OVERRIDE_RESOURCE_NAME=nvidia.com/GRID_T4-2B
VGPU_OVERRIDE_MAX_INSTANCES=8

vgpu_profile_check_command() {
    local expected_available=${1:-}
    local types=/sys/class/mdev_bus/nvidia/mdev_supported_types
    local profile="${types}/${VGPU_MDEV_TYPE}"
    local override_profile="${types}/${VGPU_OVERRIDE_MDEV_TYPE}"

    printf "test -d %q && test \"\$(cat %q)\" = %q && grep -Eq %q %q && test -d %q && test \"\$(cat %q)\" = %q && grep -Eq %q %q && test \"\$(find %q -mindepth 1 -maxdepth 1 -type d | wc -l)\" = 2" \
        "$profile" "$profile/name" "$VGPU_MDEV_NAME" \
        "max_instance=${VGPU_MAX_INSTANCES}$" "$profile/description" \
        "$override_profile" "$override_profile/name" "$VGPU_OVERRIDE_MDEV_NAME" \
        "max_instance=${VGPU_OVERRIDE_MAX_INSTANCES}$" "$override_profile/description" "$types"
    if [ -n "$expected_available" ]; then
        printf " && test \"\$(cat %q)\" = %q && test \"\$(cat %q)\" = %q" \
            "$profile/available_instances" "$expected_available" \
            "$override_profile/available_instances" "$VGPU_OVERRIDE_MAX_INSTANCES"
    fi
}
