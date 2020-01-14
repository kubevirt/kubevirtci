#!/bin/bash

set -xe

NUM_SECONDARY_NICS="${NUM_SECONDARY_NICS:-0}"

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

domain_number=1
for domain in $(virsh list --name --all); do

	# Add secondary nics
    if [ "$NUM_SECONDARY_NICS" -gt 0 ]; then
        domain_idx=$(printf "%02d" $domain_number)
        for nic_idx in $(seq -f "%02g" 1 ${NUM_SECONDARY_NICS}); do
            secondary_nic_mac=52:54:00:4b:$domain_idx:$nic_idx
            virsh attach-interface --config --model virtio --domain $domain --type network --mac $secondary_nic_mac --source $cluster_network
        done
    fi

	domain_number=$(expr $domain_number + 1)
	# Update master nodes memory
	virt-xml --edit --memory ${MASTER_MEMORY} $domain

	# Update VM's CPU mode to passthroug
	virt-xml --edit --cpu host-passthrough $domain

	# Update master nodes CPU
	[[ $domain =~ master ]] && virt-xml --edit --vcpu ${MASTER_CPU} $domain

	# Update worker nodes memory and CPU
	[[ $domain =~ worker ]] && virt-xml --edit --memory ${WORKERS_MEMORY} $domain && virt-xml --edit --vcpu ${WORKERS_CPU} $domain

	virsh start $domain

done

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

# wait for the router pod to start on the worker
until [[ $(oc -n openshift-ingress get pods -o custom-columns=NAME:.metadata.name,HOST_IP:.status.hostIP,PHASE:.status.phase | grep route | grep Running | head -n 1 | awk '{print $2}') != "" ]]; do
    sleep 5
done

# update hostnames for services to point to the node with the route pod
worker_node_ip=$(oc -n openshift-ingress get pods -o custom-columns=NAME:.metadata.name,HOST_IP:.status.hostIP,PHASE:.status.phase | grep route | grep Running | head -n 1 | awk '{print $2}')
if [[ ${worker_node_ip} != "192.168.126.51" ]]; then
    virsh net-update $cluster_network delete dns-host \
"<host ip='192.168.126.51'>
  <hostname>console-openshift-console.apps.test-1.tt.testing</hostname>
  <hostname>oauth-openshift.apps.test-1.tt.testing</hostname>
</host>" --live --config

    virsh net-update $cluster_network add dns-host \
"<host ip='${worker_node_ip}'>
  <hostname>console-openshift-console.apps.test-1.tt.testing</hostname>
  <hostname>oauth-openshift.apps.test-1.tt.testing</hostname>
</host>" --live --config

    sed -i "s/192.168.126.51/${worker_node_ip}/" /etc/haproxy/haproxy.cfg
    pkill haproxy
    haproxy -f /etc/haproxy/haproxy.cfg
fi

until [[ $(oc get pods --all-namespaces --no-headers | grep -v revision-pruner | grep -v Running | grep -v Completed | wc -l) -le 3 ]]; do
    echo "waiting for pods to come online"
    sleep 10
done

# update the pull-secret from the file
if [ -s "/etc/installer/token" ]; then
    set +x
    pull_secret=$(cat /etc/installer/token | base64 -w0)
    until oc -n openshift-config patch secret pull-secret --type merge --patch "{\"data\": {\".dockerconfigjson\": \"${pull_secret}\"}}"; do
        sleep 5
    done
    set -x
fi

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

# create local disks under all nodes, possible that we configured different number of nodes on the runtime
network_name=$(virsh net-list | grep test | awk '{print $1}')
vms=$(virsh list --name)
for vm in ${vms}; do
    vm_ip=$(virsh net-dhcp-leases ${network_name} | grep ${vm} | awk '{print $5}' | tr "/" "\t" | awk '{print $1}')
    ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -q -lcore -i vagrant.key ${vm_ip} < /scripts/create-local-disks.sh
done
