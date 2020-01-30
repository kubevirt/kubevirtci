#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR

source ../common-scripts/images.sh

../cli/cli provision --prefix k8s-fedora-${version}-provision --scripts ./scripts --k8s-version ${version} --base kubevirtci/${IMAGES[fedora31]} --tag kubevirtci/k8s-fedora-${version}
