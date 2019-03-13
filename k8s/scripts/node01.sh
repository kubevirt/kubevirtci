#!/bin/bash

set -ex

# Wait for docker, else network might not be ready yet
while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

kubeadm init --config /etc/kubernetes/kubeadm.conf

version=`kubectl version --short --client | cut -d":" -f2 |sed  's/ //g' | cut -c2- | cut -d"." -f2`

if [[ $version -ge "12" ]]; then
kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/coreos/flannel/a70459be0084506e4ec919aa1c114638878db11b/Documentation/kube-flannel.yml
else
kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/coreos/flannel/v0.9.1/Documentation/kube-flannel.yml
fi
kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/master:NoSchedule-

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
