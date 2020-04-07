#!/usr/bin/env bash

set -e

declare -A IMAGES
IMAGES[fedora31]="fedora@sha256:df2fc0d1a0db48b821c61df3e5ec2ab5575599192454656943b85dbc0d8d582a"
IMAGES[centos7]="centos@sha256:5b7a969911f2c6baf6ca8390f4da30c16dec3cacc7b5530649965c50da7240f1"
export IMAGES
