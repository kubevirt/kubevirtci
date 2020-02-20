#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

version="$(cat version | tr -d '\n')"
provision_dir="$(basename $(pwd))"

echo $version
echo $provision_dir

cd $DIR

source ../common-scripts/images.sh

export SIMPLE_PROVISION=true
../cli/cli provision --prefix k8s-${version}-provision --scripts ${provision_dir} --k8s-version ${version} --base kubevirtci/${IMAGES[centos7]} --tag kubevirtci/k8s-${provision_dir}
./check-cluster-up.sh $provision_dir
