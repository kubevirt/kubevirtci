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

../cli/cli provision --prefix os-${version}-provision --scripts ${provision_dir} --base kubevirtci/${base} --tag kubevirtci/os-${provision_dir}
