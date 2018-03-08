#!/bin/bash

set -e

docker run --privileged --rm -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/cli:latest rm
