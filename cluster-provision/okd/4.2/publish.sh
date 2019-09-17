#!/bin/bash

docker tag kubevirtci/okd-4.2-provision:latest docker.io/kubevirtci/okd-4.2:latest
docker push docker.io/kubevirtci/okd-4.2:latest
