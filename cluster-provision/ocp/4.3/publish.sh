#!/bin/bash

docker tag kubevirtci/ocp-4.3-provision:latest docker-registry.upshift.redhat.com/kubevirtci/ocp-4.3:latest
docker push docker-registry.upshift.redhat.com/kubevirtci/ocp-4.3:latest
