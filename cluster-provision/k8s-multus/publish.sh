#!/bin/bash

docker tag kubevirtci/k8s-multus-${version}:latest docker.io/kubevirtci/k8s-multus-${version}:latest
docker push docker.io/kubevirtci/k8s-multus-${version}:latest
