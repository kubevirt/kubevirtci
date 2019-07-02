#!/bin/bash

docker tag kubevirtci/centos:1905_01 docker.io/kubevirtci/centos:1905_01
docker push docker.io/kubevirtci/centos:1905_01
