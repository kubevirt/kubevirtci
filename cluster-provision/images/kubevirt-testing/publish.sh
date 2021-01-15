#!/bin/bash

docker tag kubevirtci/kubevirt-testing:latest quay.io/kubevirtci/kubevirt-testing:latest
docker push quay.io/kubevirtci/kubevirt-testing:latest
