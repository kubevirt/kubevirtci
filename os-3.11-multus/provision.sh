#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

../cli/cli provision --prefix os-3.11-multus-provision --memory 5120M --scripts ./scripts --base kubevirtci/centos@sha256:70653d952edfb8002ab8efe9581d01960ccf21bb965a9b4de4775c8fbceaab39 --tag kubevirtci/os-3.11.0-multus
