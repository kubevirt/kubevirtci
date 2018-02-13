#!/bin/bash

set -e
set -o pipefail

curl $1 | tar -zxvf - box.img
