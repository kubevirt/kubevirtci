#!/bin/bash

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR
../cli/cli provision --prefix k8s-genie-${version}-provision --scripts ./scripts --k8s-version ${version} --base kubevirtci/centos@sha256:70653d952edfb8002ab8efe9581d01960ccf21bb965a9b4de4775c8fbceaab39 --tag kubevirtci/k8s-genie-${version}
