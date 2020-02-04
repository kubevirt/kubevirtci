#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR

source ../common-scripts/images.sh

../cli/cli provision --prefix os-3.11-provision --scripts ./scripts --base kubevirtci/${IMAGES[centos7]} --tag kubevirtci/os-3.11.0
