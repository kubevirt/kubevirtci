#!/bin/bash

set -e

docker run --privileged --rm -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli:latest run --nodes 2 --base rmohr/kubeadm-1.9.3
