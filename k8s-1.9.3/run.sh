#!/bin/bash

set -e

docker run --privileged --rm -v /var/run/docker.sock:/var/run/docker.sock kubevirtci/cli:latest run --nodes 2 --background --registry-port 5000 --base kubevirtci/k8s-1.9.3
