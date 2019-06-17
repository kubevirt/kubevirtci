#!/bin/bash

set -ex

# Wait for docker, else network might not be ready yet
while [[ `systemctl status docker | grep active | wc -l` -eq 0 ]]
do
    sleep 2
done

kubeadm init --config /etc/kubernetes/kubeadm.conf
version=`kubectl version --short --client | cut -d":" -f2 |sed  's/ //g' | cut -c2- | cut -d"." -f2`

if [[ ${version} -ge "12" ]]; then
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/flannel-ge-12.yaml
else
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/flannel.yaml
fi

kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/master:NoSchedule-

kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/cna/namespace.yaml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/cna/network-addons-config.crd.yaml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/cna/operator.yaml
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/cna/network-addons-config-example.cr.yaml

# Wait until flannel cluster-network-addons-operator and core dns pods are running.
# Wait until all the network components are ready.
kubectl --kubeconfig=/etc/kubernetes/admin.conf wait networkaddonsconfig cluster --for condition=Ready --timeout=800s

kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /tmp/kubernetes-ovs-cni.yaml

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

# Setup Linux bridge used in functional tests on nodes. This won't be needed once we
# finish kubernetes-nmstate and use it for the setup instead.
./cluster-up/kubectl.sh apply -f /tmp/linux-bridge-config-iptables.yaml
while [ "$(./cluster-up/kubectl.sh get ds linux-bridge-config -o 'jsonpath={.status.numberDesiredNumberScheduled}')" == "0" ]; do sleep 1; done
while [ "$(./cluster-up/kubectl.sh get ds linux-bridge-config -o 'jsonpath={.status.numberReady}')" != "$(./cluster-up/kubectl.sh get ds linux-bridge-config -o 'jsonpath={.status.desiredNumberScheduled}')" ]; do sleep 1; done
