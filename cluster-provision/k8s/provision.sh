#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR
../cli/cli provision --prefix k8s-${version}-provision --scripts ./scripts --k8s-version ${version} --base kubevirtci/centos@sha256:4b292b646f382d986c75a2be8ec49119a03467fe26dccc3a0886eb9e6e38c911 --tag kubevirtci/k8s-${version}
