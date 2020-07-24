#!/bin/env bash

set -e

cni_name=$1
default_cidr="192.168.0.0/16"

sed -i -e "s?$default_cidr?$pod_cidr?g" $cni_name/*.yaml
