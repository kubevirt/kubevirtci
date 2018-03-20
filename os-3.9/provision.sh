#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

../cli/cli provision --scripts ./scripts --base kubevirtci/centos@sha256:94268aff21bb3b02f176b6ccbef0576d466ad31a540ca7269d6f99d31464081a --tag kubevirtci/os-3.9
