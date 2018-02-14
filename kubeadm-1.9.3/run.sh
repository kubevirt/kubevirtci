#!/bin/bash

set -e

docker run --rm -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli:latest --nodes 2 --base test
