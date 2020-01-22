#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd $DIR

source ../images.sh

../cli/cli provision --prefix os-3.11-crio-provision --crio --scripts ../os-3.11/scripts --base kubevirtci/${IMAGES[centos7]} --tag kubevirtci/os-3.11.0-crio
