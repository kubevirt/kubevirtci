#!/bin/bash

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

export version=1.17.0

cd $DIR
../provision.sh
../check-cluster-up.sh
