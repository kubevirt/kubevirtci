#!/bin/bash
set -xe

# Deploy common snapshot controller and CRDs
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/snapshot.storage.k8s.io_volumesnapshots.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/snapshot.storage.k8s.io_volumesnapshotcontents.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/snapshot.storage.k8s.io_volumesnapshotclasses.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/rbac-snapshot-controller.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/setup-snapshot-controller.yaml

# Deploy Rook/Ceph operator
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/common.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/crds.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/operator.yaml

# Create cluster
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/cluster-test.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/pool-test.yaml

until kubectl --kubeconfig /etc/kubernetes/admin.conf get cephblockpools -n rook-ceph replicapool -o jsonpath='{.status.phase}' | grep Ready; do
    ((count++)) && ((count == 120)) && echo "Ceph not ready in time" && exit 1
    if ! ((count % 6 )); then
        kubectl --kubeconfig /etc/kubernetes/admin.conf get pods -n rook-ceph
    fi
    echo "Waiting for Ceph to be Ready, sleeping 5s and rechecking"
    sleep 5
done

# k8s resources
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/storageclass-test.yaml
kubectl --kubeconfig /etc/kubernetes/admin.conf create -f /tmp/ceph/snapshotclass.yaml

# set default storageclass
kubectl --kubeconfig /etc/kubernetes/admin.conf patch storageclass local -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
kubectl --kubeconfig /etc/kubernetes/admin.conf patch storageclass rook-ceph-block -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
