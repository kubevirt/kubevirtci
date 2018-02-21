#!/bin/bash

set -e

kubeadm init --config <(cat <<EOF)
  apiVersion: kubeadm.k8s.io/v1alpha1
  kind: MasterConfiguration
  apiServerExtraArgs:
    runtime-config: admissionregistration.k8s.io/v1alpha1
  token: abcdef.1234567890123456
  pod-network-cidr: 10.244.0.0/16
  kubernetes-version: v1.9.3
EOF

kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/coreos/flannel/v0.9.1/Documentation/kube-flannel.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/master:NoSchedule-
