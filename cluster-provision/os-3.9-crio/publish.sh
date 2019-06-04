#!/bin/bash

docker tag kubevirtci/os-3.9.0-crio:latest docker.io/kubevirtci/os-3.9.0-crio:latest
docker push docker.io/kubevirtci/os-3.9.0-crio:latest
