#!/bin/bash

set -x

inventory_file="/root/inventory"
openshift_ansible="/root/openshift-ansible"

# Update inventory
echo "[new_nodes]" >> $inventory_file
sed -i '/\[OSEv3:children\]/a new_nodes' $inventory_file

nodes_found="false"
for i in $(seq 2 100); do
  node=$(printf "node%02d" ${i})
  node_ip=$(printf "192.168.66.1%02d" ${i})
  set +e
  ping ${node_ip} -c 1
  if [ $? -ne 0 ]; then
      break
  fi
  nodes_found="true"
  set -e
  echo "$node_ip $node" >> /etc/hosts
  echo "Found ${node}. Adding it to the inventory."
  echo "${node} openshift_node_group_name=\"node-config-compute\" openshift_schedulable=true openshift_ip=$node_ip" >> $inventory_file
done

# Run playbook if extra nodes were discovered
if [ "$nodes_found" = "true"  ]; then
  ansible-playbook -i $inventory_file $openshift_ansible/playbooks/openshift-node/scaleup.yml
fi

# Wait for api server to be up.
set -x
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

/usr/bin/oc create -f /tmp/local-volume.yaml

