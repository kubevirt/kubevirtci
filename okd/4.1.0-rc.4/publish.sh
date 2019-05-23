#!/bin/bash

docker tag kubevirtci/okd-4.1.0-rc.0:latest docker.io/kubevirtci/okd-4.1.0-rc.4:latest
docker push docker.io/kubevirtci/okd-4.1.0-rc.4:latest
