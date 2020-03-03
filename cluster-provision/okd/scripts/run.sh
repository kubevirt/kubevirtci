#!/bin/bash

set -xe

NUM_SECONDARY_NICS="${NUM_SECONDARY_NICS:-0}"

function oc_retry {
    until oc $@
    do
        sleep 1
    done
}

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

# update bashrc to make life easier
echo "" >> /root/.bashrc
echo 'export KUBECONFIG=/root/install/auth/kubeconfig' >> /root/.bashrc
echo "alias podc=\"oc get pods -A | grep -ivE 'run|comp'\"" >> /root/.bashrc
echo "alias pods=\"oc get pods -A\"" >> /root/.bashrc
echo "alias nodes=\"oc get nodes\"" >> /root/.bashrc
echo "alias podcw=\"oc get pods -A -owide | grep -ivE 'run|comp'\"" >> /root/.bashrc

oc_retry config set-cluster test-1 --server=https://127.0.0.1:6443
oc_retry config set-cluster test-1 --insecure-skip-tls-verify=true

# Wait for API server to be up
oc_retry get nodes

# wait half minute, just to be sure that we do not get old cluster state
sleep 30

# wait for the router pod to start on the worker
until [[ $(oc -n openshift-ingress get pods -o custom-columns=NAME:.metadata.name,HOST_IP:.status.hostIP,PHASE:.status.phase | grep route | grep Running | head -n 1 | awk '{print $2}') != "" ]]; do
    sleep 5
done

# get_value fetches command output, with retry and timeout
# syntax nodes=$(get_value 10 oc get nodes)
# first parameter is the number of iterations to try, each has 6 seconds delay
function get_value()
{
    local val=""
    timeout="$1"
    shift

    n=0
    val=$("$@")
    until [[ ${val} != "" ]]; do
        sleep 6
        n=$[$n+1]

        if [ "$n" -ge "$timeout" ]; then
            break
        fi

        val=$("$@")
    done

    echo "$val"
}

worker_node_ip=$(get_value 50 oc -n openshift-ingress get pods -o custom-columns=NAME:.metadata.name,HOST_IP:.status.hostIP,PHASE:.status.phase | grep route | grep Running | head -n 1 | awk '{print $2}')
if [[ ${worker_node_ip} == "" ]]; then
    echo "Failed to get worker_node_ip, exiting"
    exit 1
fi

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

set +xe
n=0
# Following while should iterate as long as more than 3 pods arent Ready.
# we use /tmp/num_pods.txt because we need to check NUM_PODS just in case the
# oc command itself succeeded (else value will be a fake zero).
# /tmp/timeout.inject is just optional and can be used to shrink or extend the timeout.
while true; do
    # get number of pods, when all but 3 pods are ready, continue
    oc get pods --all-namespaces --no-headers > /tmp/num_pods.txt
    if [ $? -eq 0 ]; then
        NUM_PODS=$(cat /tmp/num_pods.txt | grep -v revision-pruner | grep -v Running | grep -v Completed | wc -l)
        if [ $NUM_PODS -le 3 ] && [ $n -ge 20 ]; then
            echo $NUM_PODS "pods are not Ready, continuing cluster-up"
            break
        fi
    fi

    echo "Num of not ready pods" $NUM_PODS", waiting for pods to come up, cycle" $n
    sleep 10

    # allow to override timeout by echo timeout to timeout.inject in the container
    TIMEOUT_FILE=/tmp/timeout.inject
    timeout=90
    RE='^[0-9]+$'
    if [ -f "$TIMEOUT_FILE" ]; then
        input=$(cat $TIMEOUT_FILE)
        if [[ $input =~ $RE ]]; then
            timeout=$input
            echo "$TIMEOUT_FILE exist, overriding timeout to $timeout"
        fi
    fi

    # check if loop timeout occured
    n=$[$n+1]
    if [ "$n" -gt "$timeout" ]; then
        echo "Warning: timeout waiting for pods to come up"
        break
    fi
done
set -xe

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

echo "wait until all worker nodes will be ready"
until [[ $(oc get nodes | grep worker | grep -w Ready | wc -l) == ${WORKERS} ]]; do
    sleep 5
done

echo "wait until all master nodes will be ready"
until [[ $(oc get nodes | grep master | grep -w Ready | wc -l) == $(oc get nodes | grep master | wc -l) ]]; do
    sleep 5
done

# create local disks under all nodes, possible that we configured different number of nodes on the runtime
network_name=$(virsh net-list | grep test | awk '{print $1}')
vms=$(virsh list --name)
for vm in ${vms}; do
    vm_ip=$(virsh net-dhcp-leases ${network_name} | grep ${vm} | awk '{print $5}' | tr "/" "\t" | awk '{print $1}')
    ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -q -lcore -i vagrant.key ${vm_ip} < /scripts/create-local-disks.sh
done

set +xe
echo "cluster non ready pods:"
timeout 1m bash -c "until oc get pods -A | grep -ivE 'run|comp'; do sleep 1; done"
echo "cluster nodes status:"
timeout 1m bash -c "until oc get nodes; do sleep 1; done"
echo "NOTE: check pods state, in case it doesnt converge in reasonable time, try to restart nodes / kubelets"
