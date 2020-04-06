#!/usr/bin/env bash

set -e

declare -A IMAGES
IMAGES[fedora31]="fedora@sha256:242b0170c79034f4d2a1eadd978a6659b9761724b12aa84dadcf89a023f015cb"
IMAGES[centos7]="centos@sha256:4f105ae5eb0aa3bb034fa24591ea80bf67701b1bc6449d40aa1bec33bc6a4e9a"
export IMAGES
