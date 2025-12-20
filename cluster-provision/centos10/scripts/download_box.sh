#!/bin/bash

set -e
set -o pipefail

curl -L --fail $1 -o box.qcow2
