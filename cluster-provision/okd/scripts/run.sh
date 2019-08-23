#!/bin/bash

set -xe

# set KVM device permissions
chown root:kvm /dev/kvm
chmod 660 /dev/kvm

haproxy -f /etc/haproxy/haproxy.cfg

until virsh list
do
    sleep 5
done

cluster_network=$(virsh net-list --name | grep -v default)
# Add registry, nfs and ceph to libvirt DNS configuration
virsh net-update $cluster_network add dns-host \
"<host ip='192.168.126.1'>
  <hostname>ceph</hostname>
  <hostname>nfs</hostname>
  <hostname>registry</hostname>
</host>" --live --config

# Update VM's CPU mode to passthrough
virsh list --name --all | xargs --max-args=1 virt-xml --edit --cpu host-passthrough

# Update master nodes memory
virsh list --name --all | grep master | xargs --max-args=1 virt-xml --edit --memory ${MASTER_MEMORY}

# Update master nodes CPU
virsh list --name --all | grep master | xargs --max-args=1 virt-xml --edit --vcpu ${MASTER_CPU}

# Update worker nodes memory
virsh list --name --all | grep worker | xargs --max-args=1 virt-xml --edit --memory ${WORKERS_MEMORY}

# Update worker nodes CPU
virsh list --name --all | grep worker | xargs --max-args=1 virt-xml --edit --vcpu ${WORKERS_CPU}

# Start all VM's
virsh list --name --all | xargs --max-args=1 virsh start

while [[ "$(virsh list --name --all)" != "$(virsh list --name)" ]]; do
    sleep 1
done

export KUBECONFIG=/root/install/auth/kubeconfig
oc config set-cluster test-1 --server=https://127.0.0.1:6443
oc config set-cluster test-1 --insecure-skip-tls-verify=true

# Wait for API server to be up
until oc get nodes
do
    sleep 5
done

# wait half minute, just to be sure that we do not get old cluster state
sleep 30

until [[ $(oc get pods --all-namespaces --no-headers | grep -v Running | grep -v Completed | wc -l) -le 3 ]]; do
    echo "waiting for pods to come online"
    sleep 10
done

# update worker machine set with desired number of CPU and memory
worker_machine_set=$(oc -n openshift-machine-api get machineset --no-headers | grep worker | awk '{print $1}')
until oc -n openshift-machine-api patch machineset ${worker_machine_set} --type merge --patch "{\"spec\": {\"template\": {\"spec\": {\"providerSpec\": {\"value\": {\"domainMemory\": ${WORKERS_MEMORY}, \"domainVcpu\": ${WORKERS_CPU}}}}}}}"; do
    worker_machine_set=$(oc -n openshift-machine-api get machineset --no-headers | grep worker | awk '{print $1}')
    sleep 5
done

# update number of workers
until oc -n openshift-machine-api scale --replicas=${WORKERS} machineset ${worker_machine_set}; do
    sleep 5
done

# wait until all worker nodes will be ready
until [[ $(oc get nodes | grep worker | grep Ready | wc -l) == ${WORKERS} ]]; do
    sleep 5
done
