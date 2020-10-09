#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
provision_dir="$(basename $(pwd))"
export base="$(cat base | tr -d '\n')"

cd $DIR

(cd ../${base} && ./build.sh)
make -C ../gocli cli
../gocli/build/cli provision ${provision_dir}
./check-cluster-up.sh ${provision_dir}