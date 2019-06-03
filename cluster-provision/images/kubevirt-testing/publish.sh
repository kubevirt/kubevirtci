#!/bin/bash

docker tag kubevirtci/kubevirt-testing:latest docker.io/kubevirtci/kubevirt-testing:latest
docker push docker.io/kubevirtci/kubevirt-testing:latest
