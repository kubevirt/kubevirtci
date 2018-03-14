#!/bin/bash

# Wait until cluster will be up
set +e
/usr/local/bin/oc get nodes
while [ $? -ne 0 ]; do
    sleep 5
    /usr/local/bin/oc get nodes
done
set -e

# Update DHCP lease, will also update DNS servers
dhclient

openshift_ansible_dir="/root/openshift-ansible"
inventory_file="/root/inventory"

# Update inventory
echo "[new_nodes]" >> $inventory_file
sed -i '/\[OSEv3:children\]/a new_nodes' $inventory_file

nodes_found="false"
for i in $(seq 2 100); do
  node=$(printf "node%02d" ${i})
  num=$(printf "%02d" ${i})
  set +e
  ping -4 ${node} -c 1 
  if [ $? -ne 0 ]; then
      break
  fi
  nodes_found="true"
  set -e
  echo "Found ${node}. Adding it to the inventory."
  echo "${node} openshift_node_labels=\"{'region': 'infra','zone': 'default'}\" openshift_schedulable=true openshift_ip=192.168.66.1${num}" >> $inventory_file
done 

# Run playbook if extra nodes were discovered
if [ "$nodes_found" = "true"  ]; then
  ansible-playbook -i $inventory_file $openshift_ansible_dir/playbooks/openshift-node/scaleup.yml
fi
