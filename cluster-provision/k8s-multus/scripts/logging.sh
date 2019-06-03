#!/bin/bash

set -xe

kubectl --kubeconfig /etc/kubernetes/admin.conf create namespace logging
kubectl apply -f /tmp/logging.yaml
