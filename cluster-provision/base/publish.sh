#!/bin/bash

docker tag kubevirtci/base:latest docker.io/kubevirtci/base:latest
docker push docker.io/kubevirtci/base:latest
