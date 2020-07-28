#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

major_minor="$(basename $(pwd))"

organization=${KUBEVIRTCI_ORG:-kubevirtci}
registry=${KUBEVIRTCI_REGISTRY:-docker.io}
dryrun=${KUBEVIRTCI_DRYRUN:-false}

version="$(cat version | tr -d '\n')"
docker tag kubevirtci/k8s-${major_minor}:latest ${registry}/${organization}/k8s-${major_minor}:latest
pushcmd="docker push ${registry}/${organization}/k8s-${major_minor}:latest"


if [ "${dryrun}" == "true" ]; then
    echo dryrun: skipping provider pushing
    echo $pushcmd
else
    $pushcmd
fi
