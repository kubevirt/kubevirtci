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

# Deploy Ceph secret and storageclass
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/ceph-secret.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/ceph-storageclass.yaml
