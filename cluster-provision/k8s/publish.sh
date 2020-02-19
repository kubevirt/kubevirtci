#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

version="$(cat version | tr -d '\n')"
docker tag kubevirtci/k8s-${version}:latest docker.io/kubevirtci/k8s-${version}:latest
docker push docker.io/kubevirtci/k8s-${version}:latest
