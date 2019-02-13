#!/bin/bash

set -x

inventory_file="/root/inventory"
openshift_ansible="/root/openshift-ansible"

# Update inventory
nodes_found=0
for i in $(seq 2 100); do
  node=$(printf "node%02d" ${i})
  node_ip=$(printf "192.168.66.1%02d" ${i})
  set +e
  ping ${node_ip} -c 1
  if [ $? -ne 0 ]; then
      break
  fi
  echo "Found ${node}. Adding it to inventory and hosts files."
  # add after first "hosts:" line
  sed -i "0,/hosts:/{s//hosts:\n        ${node}:\n          openshift_ip: $node_ip\n          openshift_node_group_name: node-config-compute-kubevirt\n          openshift_schedulable: true/}" $inventory_file
  echo "$node_ip $node" >> /etc/hosts
  let "nodes_found++"
done

# Run playbook if extra nodes were discovered
if [ "$nodes_found" -gt 0 ]; then
  # first modify inventory: add new_nodes to OSEv3 children (which is the first children line prefixed with 6 spaces)
  let last_node_nr=nodes_found+1
  last_node_nr=$(printf "%02d" ${last_node_nr})
  sed -i "0,/      children:/{s//      children:\n        new_nodes:\n          hosts:\n            node[02:${last_node_nr}]:/}" $inventory_file
  ansible-playbook -i $inventory_file $openshift_ansible/playbooks/openshift-node/scaleup.yml
fi

set +e
crio=false
grep crio $inventory_file
if [ $? -eq 0 ]; then
  crio=true
fi
set -e

cat >post_deployment_configuration <<EOF
- hosts: nodes, new_nodes
  tasks:
    - name: Configure CRI-O support
      block:
        - replace:
            path: /etc/crio/crio.conf
            regexp: 'insecure_registries = \[\n""\n\]'
            replace: 'insecure_registries = ["docker.io", "registry:5000"]'
        - replace:
            path: /etc/crio/crio.conf
            regexp: 'registries = \[\n"docker.io"\n\]'
            replace: 'registries = ["docker.io", "registry:5000"]'
        - service:
            name: cri-o
            state: restarted
            enabled: yes
      when: crio
    - name: Clean cpu manager state
      block:
        - file:
            state: absent
            path: /var/lib/origin/openshift.local.volumes/cpu_manager_state
        - service:
            name: origin-node
            state: restarted
            enabled: yes

EOF
ansible-playbook -i $inventory_file post_deployment_configuration --extra-vars="crio=${crio}"

# Wait for api server to be up.
set -x
set +e
/usr/bin/oc get nodes --no-headers
os_rc=$?
retry_counter=0
while [[ $retry_counter -lt 20  && $os_rc -ne 0 ]]; do
    sleep 10
    echo "Waiting for api server to be available..."
    /usr/bin/oc get nodes --no-headers
    os_rc=$?
    retry_counter=$((retry_counter + 1))
done
set -e

# Remove the multus pod to recreate it.
# openshift-sdn remove the content of the cni folder on restart.
# Needs to recreate the multus cni config after the openshift-sdn is up.
oc -n kube-system delete po `oc get po -n kube-system | grep kube-multus-ds | awk '{print $1}'`
oc -n kube-system delete po `oc get po -n kube-system | grep kube-cni-plugins | awk '{print $1}'`
oc -n kube-system delete po `oc get po -n kube-system | grep ovs-cni | awk '{print $1}'`

/usr/bin/oc create -f /tmp/local-volume.yaml
