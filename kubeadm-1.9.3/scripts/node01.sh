#!/bin/bash

set -e

kubeadm init --pod-network-cidr=10.244.0.0/16 --kubernetes-version v1.9.3 --token abcdef.1234567890123456
kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/coreos/flannel/v0.9.1/Documentation/kube-flannel.yml
