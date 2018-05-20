#!/bin/bash

set -ex

cat > /etc/kubernetes/kubeadm.conf <<EOF
apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
apiServerExtraArgs:
  runtime-config: admissionregistration.k8s.io/v1alpha1
  admission-control: Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota
token: abcdef.1234567890123456
kubernetesVersion: v1.9.3
networking:
  podSubnet: 10.244.0.0/16
EOF

kubeadm init --config /etc/kubernetes/kubeadm.conf

kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/coreos/flannel/v0.9.1/Documentation/kube-flannel.yml
kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/master:NoSchedule-
