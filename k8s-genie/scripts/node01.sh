#!/bin/bash

set -ex

# Wait for docker, else network might not be ready yet
while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

kubeadm init --config /etc/kubernetes/kubeadm.conf

kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/genie.yaml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/flannel.yaml
kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/master:NoSchedule-

# update the genie configuration to use flannel as default cni plugin
kubectl --kubeconfig=/etc/kubernetes/admin.conf replace -f /tmp/genie-configmap.yaml

# Wait for api server to be up.
kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers
kubectl_rc=$?
retry_counter=0
while [[ $retry_counter -lt 20 && $kubectl_rc -ne 0 ]]; do
    sleep 10
    echo "Waiting for api server to be available..."
    kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers
    kubectl_rc=$?
    retry_counter=$((retry_counter + 1))
done
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/local-volume.yaml

# set ptp cni static configuration after genie already injected itself into that directory
set +x
echo "wait for genie to inject its configuration to /etc/cni/net.d/"
while [ "$(ls -A /etc/cni/net.d/ | wc -l)" -eq 0 ]; do
     sleep 1
done
set -x

cp /tmp/static-ptp-conf.yaml /etc/cni/net.d/10-ptp.conf
