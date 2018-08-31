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

# enable the CPU manager on compute nodes, besides master
# updating it on the master as nodes cannot update comfigmaps

sed -i 's/kubeletArguments:/kubeletArguments:\n  cpu-manager-policy:\n  - static\n  system-reserved:\n  - cpu=500m\n  kube-reserved:\n  - cpu=500m/' /etc/origin/node/node-config.yaml
sed -i 's/feature-gates:/feature-gates:\n  - CPUManager=true/' /etc/origin/node/node-config.yaml

/usr/bin/oc create cm node-config-compute -n openshift-node --from-file=/etc/origin/node/node-config.yaml -o yaml --dry-run |oc replace -f -

cat >post_deployment_cpu_manager_config <<EOF
- hosts: nodes, new_nodes
  tasks:
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

sleep 200

ansible-playbook -i $inventory_file post_deployment_cpu_manager_config

