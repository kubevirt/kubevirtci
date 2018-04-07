#!/bin/bash

set -ex

# Install storage requirements for iscsi and cluster
yum -y install centos-release-gluster
yum -y install --nogpgcheck -y glusterfs-fuse
yum -y install iscsi-initiator-utils

# Install OpenShift packages
yum install -y centos-release-openshift-origin
yum install -y yum-utils ansible wget git net-tools bind-utils iptables-services bridge-utils bash-completion kexec-tools sos psacct docker-1.12.6-71.git3e8e77d.el7.centos

# Disable spectre and meltdown patches
sed -i 's/quiet"/quiet spectre_v2=off nopti"/' /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

echo 0 > /sys/kernel/debug/x86/pti_enabled
echo 0 > /sys/kernel/debug/x86/ibpb_enabled
echo 0 > /sys/kernel/debug/x86/ibrs_enabled

sed -i 's/--log-driver=journald //g' /etc/sysconfig/docker
echo '{ "insecure-registries" : ["registry:5000"] }' > /etc/docker/daemon.json

systemctl start docker
systemctl enable docker

# Allow connecting to ssh via password
sed -i -e "s/PasswordAuthentication no/PasswordAuthentication yes/" /etc/ssh/sshd_config
systemctl restart sshd

# Disable host key checking under ansible.cfg file
sed -i '/host_key_checking/s/^#//g' /etc/ansible/ansible.cfg

openshift_ansible_dir="/root/openshift-ansible"
inventory_file="/root/inventory"
master_ip="192.168.66.101"
echo "$master_ip node01" >> /etc/hosts 

mkdir -p /root/openshift-ansible
# Checkout to the specific version as W/A for https://github.com/openshift/openshift-ansible/issues/6756
git clone https://github.com/openshift/openshift-ansible.git $openshift_ansible_dir -b openshift-ansible-3.9.0-0.42.0

# Create ansible inventory file
cat >$inventory_file <<EOF
[OSEv3:children]
masters
nodes

[OSEv3:vars]
ansible_ssh_user=root
ansible_ssh_pass=vagrant
deployment_type=origin
openshift_deployment_type=origin
openshift_clock_enabled=true
openshift_master_identity_providers=[{'name': 'allow_all_auth', 'login': 'true', 'challenge': 'true', 'kind': 'AllowAllPasswordIdentityProvider'}]
openshift_disable_check=memory_availability,disk_availability,docker_storage,package_availability,docker_image_availability
openshift_repos_enable_testing=True
openshift_image_tag=v3.9.0-alpha.4
containerized=true
enable_excluders=false
ansible_service_broker_registry_whitelist=['.*-apb$']
openshift_hosted_etcd_storage_kind=nfs
openshift_hosted_etcd_storage_nfs_options="*(rw,root_squash,sync,no_wdelay)"
openshift_hosted_etcd_storage_nfs_directory=/opt/etcd-vol
openshift_hosted_etcd_storage_volume_name=etcd-vol
openshift_hosted_etcd_storage_access_modes=["ReadWriteOnce"]
openshift_hosted_etcd_storage_volume_size=1G
openshift_hosted_etcd_storage_labels={'storage': 'etcd'}
openshift_node_kubelet_args={'max-pods': ['40'], 'pods-per-core': ['40']}

[nfs]
node01 openshift_ip=$master_ip

[masters]
node01 openshift_ip=$master_ip

[etcd]
node01 openshift_ip=$master_ip

[nodes]
node01 openshift_node_labels="{'region': 'infra','zone': 'default'}" openshift_schedulable=true openshift_ip=$master_ip
EOF

# Run OpenShift ansible playbook
ansible-playbook -e "ansible_user=root ansible_ssh_pass=vagrant" -i $inventory_file $openshift_ansible_dir/playbooks/prerequisites.yml
ansible-playbook -i $inventory_file $openshift_ansible_dir/playbooks/deploy_cluster.yml

# Create OpenShift user
/usr/local/bin/oc create user admin
/usr/local/bin/oc create identity allow_all_auth:admin
/usr/local/bin/oc create useridentitymapping allow_all_auth:admin admin
/usr/local/bin/oc adm policy add-cluster-role-to-user cluster-admin admin
