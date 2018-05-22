#!/bin/bash

docker tag kubevirtci/centos:1804_02 docker.io/kubevirtci/centos:1804_02
docker push docker.io/kubevirtci/centos:1804_02
