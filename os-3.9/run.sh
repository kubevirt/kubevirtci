#!/bin/bash

set -e

../cli/cli run --nodes 2 --reverse --background --memory 4096M --base kubevirtci/os-3.9
