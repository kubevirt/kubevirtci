#!/bin/bash
okd_version="4.1.0-rc.4"

docker tag kubevirtci/okd-${okd_version}:latest docker.io/kubevirtci/okd-${okd_version}:latest
docker push docker.io/kubevirtci/okd-${okd_version}:latest
