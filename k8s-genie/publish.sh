#!/bin/bash

docker tag kubevirtci/k8s-genie-${version}:latest docker.io/kubevirtci/k8s-genie-${version}:latest
docker push docker.io/kubevirtci/k8s-genie-${version}:latest
