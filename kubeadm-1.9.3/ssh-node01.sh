#!/bin/bash

set -e

docker run --privileged --rm -it -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli:latest ssh node01 
