#!/usr/bin/env bash
set -exuo pipefail

SCRIPT_PATH=$(dirname "$(realpath "$0")")
archs=(aarch64 x86_64)

main() {
    local build_only=""
    while getopts "nbh" opt; do
        case "$opt" in
            n)
		archs=("$(uname -m)")
                ;;
            b)
                build_only=true
                ;;
            h)
                help
                exit 0
                ;;
            *)
                echo "Invalid argument: $opt"
                help
                exit 1
        esac
    done
    shift $((OPTIND-1))
    local build_target="${1:?}"
    local registry="${2:?}"
    local registry_org="${3:?}"
    local full_image_name image_tag base_image
    local tag=${KUBEVIRTCI_TAG:-"$(get_image_tag)"}

    full_image_name="$(
        get_full_image_name \
            "$registry" \
            "$registry_org" \
            "${build_target##*/}" \
            "${tag}"
    )"

    build_image "$build_target"

    [[ $build_only ]] && return
    publish_image "$full_image_name"
    publish_manifest "$full_image_name"
}

help() {
    cat <<EOF
    Usage:
        ./publish_multiarch_image.sh [OPTIONS] BUILD_TARGET REGISTRY REGISTRY_ORG

    Build and publish multiarch infra images.

    OPTIONS
        -n  (native build) Only build image for host CPU Arch.
        -h  Show this help message and exit.
        -b  Only build the image and exit. Do not publish the built image.
EOF
}

get_image_tag() {
    local current_commit today
    current_commit="$(git rev-parse HEAD)"
    today="$(date +%Y%m%d)"
    echo "v${today}-${current_commit:0:7}"
}

get_full_image_name() {
    local registry="${1:?}"
    local registry_org="${2:?}"
    local image_name="${3:?}"
    local tag="${4:?}"

    echo "${registry_org}/${registry}/${image_name}:${tag}"
}

build_image() {
    local build_target="${1:?}"
    # build multi-arch images
    for arch in ${archs[*]};do
        ARCHITECTURE=${arch} ${SCRIPT_PATH}/create-containerdisk.sh ${build_target}
    done
}

publish_image() {
    local full_image_name="${1:?}"
    for arch in ${archs[*]};do
        docker tag ${build_target}:devel-${arch} ${full_image_name}-${arch}
        skopeo copy "docker-daemon:${full_image_name}-${arch}" "docker://${full_image_name}-${arch}"
    done
}

publish_manifest() {
    local amend
    local full_image_name="${1:?}"
    amend=""
    for arch in ${archs[*]};do
        amend+=" --amend ${full_image_name}-${arch}"
    done
    docker manifest create ${full_image_name} ${amend}
    docker manifest push ${full_image_name} "docker://${full_image_name}"
}

main "$@"
