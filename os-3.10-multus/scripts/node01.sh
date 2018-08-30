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
EOF
ansible-playbook -i $inventory_file post_deployment_configuration --extra-vars="crio=${crio}"

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

/usr/bin/oc create -f /tmp/multus.yaml
/usr/bin/oc create -f /tmp/macvlan-conf.yaml
/usr/bin/oc create -f /tmp/ptp-conf.yaml
/usr/bin/oc create -f /tmp/bridge-conf.yaml

# create local block device, backed by raw cirros disk image (see also provision.sh)
LOOP_DEVICE=`losetup --find --show /mnt/local-storage/cirros.img.raw`
rm -f /mnt/local-storage/cirros-block-device
ln -s $LOOP_DEVICE /mnt/local-storage/cirros-block-device
