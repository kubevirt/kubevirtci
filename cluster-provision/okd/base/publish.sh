#!/bin/bash

docker tag kubevirtci/okd-base docker.io/kubevirtci/okd-base
docker push docker.io/kubevirtci/okd-base
