#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

../cli/cli provision --scripts ./scripts --base kubevirtci/centos@sha256:eeacdb20f0f5ec4e91756b99b9aa3e19287a6062bab5c3a41083cd245a44dc43 --tag kubevirtci/os-3.9
