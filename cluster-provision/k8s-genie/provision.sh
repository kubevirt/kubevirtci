#!/bin/bash

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR
../cli/cli provision --prefix k8s-genie-${version}-provision --scripts ./scripts --k8s-version ${version} --base kubevirtci/centos@sha256:f2b204a3fabf71494c3cf7a145dae88e63f068a84ca10149b0aec523045dc2c1 --tag kubevirtci/k8s-genie-${version}
