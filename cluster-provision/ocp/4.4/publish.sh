#!/bin/bash

tag=$(git log -1 --pretty=%h)-$(date +%s)
destination="quay.io/kubevirtci/ocp-4.4:$tag"

docker tag kubevirtci/ocp-4.4-provision:latest $destination
docker push $destination
