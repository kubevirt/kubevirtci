#!/bin/bash

set -e

docker run --privileged --rm -v /var/run/docker.sock:/var/run/docker.sock rmohr/cli:latest ps 
