#!/bin/bash

set -ex

kubeadm init --config /etc/kubernetes/kubeadm.conf

kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f ${cni_plugin}
kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/master:NoSchedule-
