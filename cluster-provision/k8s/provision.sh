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

(cd ../gocli && make cli)
../gocli/build/cli provision --prefix k8s-${version}-provision --scripts ${provision_dir} --k8s-version ${version} kubevirtci/${base} kubevirtci/k8s-${provision_dir}
./check-cluster-up.sh $provision_dir