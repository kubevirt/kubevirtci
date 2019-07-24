#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR
../cli/cli provision --prefix k8s-${version}-provision --scripts ./scripts --k8s-version ${version} --base kubevirtci/centos@sha256:728232c8b5b79bb69e1386a6b47acd72876faa386202b535e227bc605adffe33 --tag kubevirtci/k8s-${version}
