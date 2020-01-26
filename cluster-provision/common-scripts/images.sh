#!/usr/bin/env bash

set -e

declare -A IMAGES
IMAGES[fedora31]=""
IMAGES[fedora31-vagrant]="31-1.9"
IMAGES[centos7]="centos@sha256:6f2548dcc23489d0c945aef516781ae2ea678424c3760d1dafa0a83d29411713"
IMAGES[centos7-vagrant]="2001_01"
export IMAGES
