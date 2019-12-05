#!/bin/bash

docker tag kubevirtci/ocp-4.3-provision:latest quay.io/openshift-cnv/kubevirtci-ocp-4.3:latest
docker push quay.io/openshift-cnv/kubevirtci-ocp-4.3:latest
