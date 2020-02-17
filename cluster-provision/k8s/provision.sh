#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

base="$(cat base | tr -d '\n')"
version="$(cat version | tr -d '\n')"
provision_dir="$(basename $(pwd))"

echo $version
echo $base
echo $provision_dir

cd $DIR

export SIMPLE_PROVISION=true
gocli="$DIR/../../cluster-provision/gocli/run.sh"

$gocli provision --prefix k8s-${version}-provision --scripts $DIR/${provision_dir} --k8s-version ${version} --tag kubevirtci/k8s-${provision_dir} kubevirtci/${base}
#./check-cluster-up.sh $provision_dir
