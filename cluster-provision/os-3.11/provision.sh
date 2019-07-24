#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

../cli/cli provision --prefix os-3.11-provision --scripts ./scripts --base kubevirtci/centos@sha256:728232c8b5b79bb69e1386a6b47acd72876faa386202b535e227bc605adffe33 --tag kubevirtci/os-3.11.0
