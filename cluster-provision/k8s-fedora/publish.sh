#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

docker tag kubevirtci/k8s-fedora-${version}:latest docker.io/kubevirtci/k8s-fedora-${version}:latest
docker push docker.io/kubevirtci/k8s-fedora-${version}:latest
