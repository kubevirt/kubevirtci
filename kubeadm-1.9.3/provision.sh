#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

docker run --rm -v ${DIR}/scripts/:/scripts/ -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli --provision --scripts /scripts --base rmohr/centos:1608_01 --tag test
