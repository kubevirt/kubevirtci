#!/usr/bin/env bash

set -e

declare -A IMAGES
IMAGES[fedora31]="fedora@sha256:6e32c9c0073bd79a435537a067f14e7f9b72e1ddd9229f711306a93b9252125a"
IMAGES[centos7]="centos@sha256:6f2548dcc23489d0c945aef516781ae2ea678424c3760d1dafa0a83d29411713"
export IMAGES
