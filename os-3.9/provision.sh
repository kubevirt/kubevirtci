#!/bin/bash

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

../cli/cli provision --scripts ./scripts --base kubevirtci/centos@sha256:bd2bf287ce3b28a3624575b5dd31e375bbb213502693c4723d7a945e12dcf0f8 --tag kubevirtci/os-3.9
