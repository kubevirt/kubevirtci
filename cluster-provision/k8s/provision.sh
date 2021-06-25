#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
provision_dir="$(basename "$(pwd)")"
base="$(cat base | tr -d '\n')"
export base

cd $DIR

gocli_args=""

if [ -n "${CONTAINER_SUFFIX}" ]; then
    gocli_args="${gocli_args} --container-suffix=${CONTAINER_SUFFIX}"
fi

(cd ../${base} && ./build.sh)
make -C ../gocli cli
../gocli/build/cli provision ${gocli_args} ${provision_dir}
CONTAINER_SUFFIX=${CONTAINER_SUFFIX} ./check-cluster-up.sh ${provision_dir}
