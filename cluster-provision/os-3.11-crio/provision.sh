#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

../cli/cli provision --prefix os-3.11-crio-provision --crio --scripts ../os-3.11/scripts --base kubevirtci/centos@sha256:f2b204a3fabf71494c3cf7a145dae88e63f068a84ca10149b0aec523045dc2c1 --tag kubevirtci/os-3.11.0-crio
