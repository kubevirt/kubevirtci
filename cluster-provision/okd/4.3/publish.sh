#!/bin/bash

docker tag kubevirtci/okd-4.3-provision:latest docker.io/kubevirtci/okd-4.3:latest
docker push docker.io/kubevirtci/okd-4.3:latest
