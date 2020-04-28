#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

major_minor="$(basename $(pwd))"
tag=$(git log -1 --pretty=%h)-$(date +%s)
destination="docker.io/kubevirtci/k8s-${major_minor}:$tag"

docker tag kubevirtci/k8s-${major_minor}:latest $destination
docker push $destination
