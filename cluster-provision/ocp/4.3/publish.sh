#!/bin/bash
tag=$(git log -1 --pretty=%h)-$(date +%s)
destination="quay.io/kubevirtci/ocp-4.3:$tag"
echo docker tag kubevirtci/ocp-4.3-provision:latest $destination
echo docker push $destination
