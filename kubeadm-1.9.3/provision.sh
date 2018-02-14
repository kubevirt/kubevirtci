#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker run --privileged --rm -v ${DIR}/scripts/:/scripts/ -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli provision --scripts /scripts --base rmohr/centos:1608_01 --tag rmohr/kubeadm-1.9.3
