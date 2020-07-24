#!/bin/bash
# DO NOT RUN THIS SCRIPT, USE SCRIPTS UNDER VERSIONS DIRECTORIES

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

provision_dir="$(basename $(pwd))"

docker tag kubevirtci/os-${provision_dir}:latest docker.io/kubevirtci/os-${provision_dir}:latest
docker push docker.io/kubevirtci/os-${provision_dir}:latest
