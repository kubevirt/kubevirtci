#!/bin/bash

docker tag kubevirtci/os-3.11.0-multus-sriov:latest docker.io/kubevirtci/os-3.11.0-multus-sriov:latest
docker push docker.io/kubevirtci/os-3.11.0-multus-sriov:latest
