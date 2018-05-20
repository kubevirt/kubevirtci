#!/bin/bash

docker tag kubevirtci/k8s-1.9.3:latest docker.io/kubevirtci/k8s-1.9.3:latest
docker push docker.io/kubevirtci/k8s-1.9.3:latest
