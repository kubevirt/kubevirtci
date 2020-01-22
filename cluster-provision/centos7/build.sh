#!/bin/bash -e

centos_version=$(cat version)

docker build --build-arg centos_version=$centos_version . -t kubevirtci/centos:$centos_version
