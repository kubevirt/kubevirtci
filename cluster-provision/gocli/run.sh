#!/bin/bash

set -xe

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

make cli

build/cli $@
