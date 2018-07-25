#!/bin/bash

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR
../cli/cli provision --prefix k8s-${version}-provision --scripts ./scripts --k8s-version ${version} --base kubevirtci/centos@sha256:5539557ff8cbe96a3ef05e5705f82b58c38e1ff1cdf09f55a47aa5eb542f4ce8 --tag kubevirtci/k8s-multul-${version}
