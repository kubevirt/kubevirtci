#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
provision_dir="$(basename $(pwd))"
export base="$(cat base | tr -d '\n')"

cd $DIR

gocli_args=""

if [ "${CGROUPV2}" == "true" ]; then
    gocli_args="${gocli_args} --cgroupv2=true"
    CONTAINER_SUFFIX=cgroupsv2
fi

if [ -n "${CONTAINER_SUFFIX}" ]; then
    gocli_args="${gocli_args} --container-suffix=${CONTAINER_SUFFIX}"
fi

(cd ../${base} && ./build.sh)
make -C ../gocli cli
cp "${DIR}/fetch-images.sh" "${provision_dir}/"
trap 'rm -f ${provision_dir}/fetch-images.sh' EXIT SIGINT SIGTERM
../gocli/build/cli provision ${gocli_args} ${provision_dir}
CONTAINER_SUFFIX=${CONTAINER_SUFFIX} ./check-cluster-up.sh ${provision_dir}
