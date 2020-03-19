#!/bin/bash

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
../cli/cli provision --prefix k8s-${version}-provision --scripts ${provision_dir} --k8s-version ${version} --base kubevirtci/${base} --tag kubevirtci/k8s-${provision_dir}
./check-cluster-up.sh $provision_dir

# Adjust the sha256 in images.sh
array_item="IMAGES\[k8s-${provision_dir}\]"
docker_sha=`docker image inspect -f '{{.Id}}' kubevirtci/k8s-${provision_dir}:latest`
docker_image="k8s-${provision_dir}@${docker_sha}"
sed -i "s/${array_item}=.*/${array_item}=\"${docker_image}\"/" ../../cluster-up/cluster/images.sh
