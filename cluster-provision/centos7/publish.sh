#!/bin/bash -e

centos_version=$(cat version)

docker tag kubevirtci/centos:$centos_version docker.io/kubevirtci/centos:$centos_version
docker push docker.io/kubevirtci/centos:$centos_version
