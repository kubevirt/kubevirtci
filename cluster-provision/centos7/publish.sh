#!/bin/bash -e

docker tag kubevirtci/centos7 docker.io/kubevirtci/centos7
docker push docker.io/kubevirtci/centos7
