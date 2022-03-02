#!/bin/bash
set -xe

# Deploy NFS CSI manifests
kubectl --kubeconfig /etc/kubernetes/admin.conf create ns nfs-csi
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/nfs-csi/nfs-service.yaml -n nfs-csi
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/nfs-csi/nfs-server.yaml -n nfs-csi
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/nfs-csi/csi-nfs-controller-rbac.yaml -n nfs-csi
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/nfs-csi/csi-nfs-driverinfo.yaml -n nfs-csi
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/nfs-csi/csi-nfs-controller.yaml -n nfs-csi
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/nfs-csi/csi-nfs-node.yaml -n nfs-csi
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/nfs-csi/csi-nfs-sc.yaml -n nfs-csi

# Deploy test PVC and wait for it to get bound
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/nfs-csi/csi-nfs-test-pvc.yaml -n nfs-csi
until kubectl --kubeconfig /etc/kubernetes/admin.conf get pvc -n nfs-csi pvc-nfs-dynamic -o jsonpath='{.status.phase}' | grep Bound; do
    ((count++)) && ((count == 120)) && echo "NFS CSI test PVC not ready on time" && exit 1
    if ! ((count % 6 )); then
        kubectl --kubeconfig /etc/kubernetes/admin.conf describe pvc -n nfs-csi
    fi
    echo "Waiting for NFS CSI test PVC to be Bound, sleeping 5s and rechecking"
    sleep 5
done
kubectl --kubeconfig /etc/kubernetes/admin.conf delete -f /tmp/nfs-csi/csi-nfs-test-pvc.yaml -n nfs-csi
