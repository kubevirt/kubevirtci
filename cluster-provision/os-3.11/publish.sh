#!/bin/bash

docker tag kubevirtci/os-3.11.0:latest docker.io/kubevirtci/os-3.11.0:latest
docker push docker.io/kubevirtci/os-3.11.0:latest
