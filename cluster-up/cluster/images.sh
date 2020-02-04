#!/usr/bin/env bash

set -e

declare -A IMAGES
IMAGES[gocli]="gocli@sha256:220f55f6b1bcb3975d535948d335bd0e6b6297149a3eba1a4c14cad9ac80f80d"
if [ -z $KUBEVIRTCI_PROVISION_CHECK ]; then
    IMAGES[k8s-fedora-1.17.0]="k8s-fedora-1.17.0@sha256:f6901951bb289f91b50e5063310b72e670ba83c62111b34a8882517a4a7ee4c5"
    IMAGES[k8s-1.17.0]="k8s-1.17.0@sha256:d790b77d514f60a0e75fcce39b4161c2fe17703d52933745bdbcba20a63dc13c"
    IMAGES[k8s-1.16.2]="k8s-1.16.2@sha256:5bae6a5f3b996952c5ceb4ba12ac635146425909801df89d34a592f3d3502b0c"
    IMAGES[k8s-1.15.1]="k8s-1.15.1@sha256:14d7b1806f24e527167d2913deafd910ea46e69b830bf0b094dde35ba961b159"
    IMAGES[k8s-1.14.6]="k8s-1.14.6@sha256:ec29c07c94fce22f37a448cb85ca1fb9215d1854f52573316752d19a1c88bcb3"
    IMAGES[k8s-1.13.3]="k8s-1.13.3@sha256:afbdd9b4208e5ce2ec327f302c336cea3ed3c22488603eab63b92c3bfd36d6cd"
    IMAGES[k8s-1.11.0]="k8s-1.11.0@sha256:696ba7860fc635628e36713a2181ef72568d825f816911cf857b2555ea80a98a"
    IMAGES[k8s-genie-1.11.1]="k8s-genie-1.11.1@sha256:19af1961fdf92c08612d113a3cf7db40f02fd213113a111a0b007a4bf0f3f7e7"
    IMAGES[k8s-multus-1.13.3]="k8s-multus-1.13.3@sha256:c0bcf0d2e992e5b4d96a7bcbf988b98b64c4f5aef2f2c4d1c291e90b85529738"
    IMAGES[okd-4.1]="okd-4.1@sha256:e7e3a03bb144eb8c0be4dcd700592934856fb623d51a2b53871d69267ca51c86"
    IMAGES[okd-4.2]="okd-4.2@sha256:a830064ca7bf5c5c2f15df180f816534e669a9a038fef4919116d61eb33e84c5"
    IMAGES[okd-4.3]="okd-4.3@sha256:63abc3884002a615712dfac5f42785be864ea62006892bf8a086ccdbca8b3d38"
    IMAGES[ocp-4.3]="ocp-4.3@sha256:8f59d625852ef285d6ce3ddd6ebd3662707d2c0fab19772b61dd0aa0f6b41e5f"
    IMAGES[ocp-4.4]="ocp-4.4@sha256:b235e87323ed88c46fedf27e9115573b92f228a82559ab7523dd1be183f66af8"
    IMAGES[ocp-cnao-4.4]="ocp-cnao-4.4@sha256:b326ddb9b0d503e3ea469c442d563502ea8cce57b1fe8f8a83bd1a29bbf71c32"
fi
export IMAGES

IMAGE_SUFFIX=""
if [[ $KUBEVIRT_PROVIDER =~ (ocp|okd).* ]]; then
    IMAGE_SUFFIX="-provision"
fi

image="${IMAGES[$KUBEVIRT_PROVIDER]:-${KUBEVIRT_PROVIDER}${IMAGE_SUFFIX}:latest}"
export image
