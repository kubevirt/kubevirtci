#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

../cli/cli provision --prefix os-3.11-multus-provision --memory 5120M --scripts ./scripts --base kubevirtci/centos@sha256:4b292b646f382d986c75a2be8ec49119a03467fe26dccc3a0886eb9e6e38c911 --tag kubevirtci/os-3.11.0-multus
