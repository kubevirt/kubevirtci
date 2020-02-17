#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR

version="$(cat version | tr -d '\n')"

source ../common-scripts/images.sh

../cli/cli provision --prefix k8s-${version}-provision --scripts ./scripts --k8s-version ${version} --base kubevirtci/${IMAGES[centos7]} --tag kubevirtci/k8s-${version}
