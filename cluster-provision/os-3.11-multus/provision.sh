#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR

source ../images.sh

../cli/cli provision --prefix os-3.11-multus-provision --memory 5120M --scripts ./scripts --base kubevirtci/${IMAGES[centos7]} --tag kubevirtci/os-3.11.0-multus
