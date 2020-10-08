#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
provision_dir="$(basename $(pwd))"
cd $DIR

(cd ../gocli && make cli)
../gocli/build/cli provision ${provision_dir}
./check-cluster-up.sh ${provision_dir}