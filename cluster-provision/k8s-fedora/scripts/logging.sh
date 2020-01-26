#!/bin/bash

set -xe

kubectl --kubeconfig /etc/kubernetes/admin.conf create namespace logging
kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f /tmp/logging.yaml
