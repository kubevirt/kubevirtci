#!/bin/bash
set -xe

yum install -y ceph-common

# Deploy RBACs for sidecar containers and node plugins
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/csi-attacher-rbac.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/csi-provisioner-rbac.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/csi-nodeplugin-rbac.yaml

# Deploy CSI sidecar containers
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/csi-rbdplugin-attacher.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/csi-rbdplugin-provisioner.yaml

# Deploy RBD CSI driver
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/csi-rbdplugin.yaml

# Deploy Ceph secret, storageclass and snapshotclass
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/ceph-secret.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/ceph-storageclass.yaml

sleep 10

set +e

status=`kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods --all-namespaces --no-headers --field-selector status.phase!=Running --output=name | grep rbdplugin`
retry_counter=0
while [[ $retry_counter -lt 30 && ! -z "$status" ]]; do
    sleep 10
    echo "Waiting for rbd plugin to be available..."
    status=`kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods --all-namespaces --no-headers --field-selector status.phase!=Running --output=name | grep rbdplugin`
    retry_counter=$((retry_counter + 1))
done

set -e

if [[ -z "$status" ]]
then
    kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/ceph-snapshotclass.yaml
else
    echo "SnapshotClass not created- RBD plugin not Ready"
fi
