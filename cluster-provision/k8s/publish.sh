#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

major_minor="$(basename $(pwd))"

version="$(cat version | tr -d '\n')"
docker tag kubevirtci/k8s-${major_minor}:latest docker.io/kubevirtci/k8s-${major_minor}:latest
docker push docker.io/kubevirtci/k8s-${major_minor}:latest
