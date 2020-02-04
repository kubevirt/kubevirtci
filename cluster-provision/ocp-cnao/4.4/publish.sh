#!/bin/bash -xe

name=ocp-cnao-4.4
tag=$(git log -1 --pretty=%h)-$(date +%s)
destination="quay.io/kubevirtci/$name:$tag"

docker tag kubevirtci/$name-provision:latest $destination
docker push $destination
