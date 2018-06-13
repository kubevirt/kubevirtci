#!/bin/bash

set -ex

# Set SELinux to permissive. Still logging the denials.
setenforce 0
sed -i "s/^SELINUX=.*/SELINUX=permissive/" /etc/selinux/config

# Install epel
yum -y install epel-release

# Install storage requirements for iscsi and cluster
yum -y install centos-release-gluster
yum -y install --nogpgcheck -y glusterfs-fuse
yum -y install iscsi-initiator-utils

# Create Origin latest repo
cat >/etc/yum.repos.d/origin-latest.repo <<EOF
[centos-openshift-origin-latest]
name=CentOS OpenShift Origin Latest
baseurl=https://cbs.centos.org/repos/paas7-openshift-origin39-candidate/x86_64/os/
enabled=1
gpgcheck=0
EOF

# Install OpenShift packages
yum install -y centos-release-openshift-origin
yum install -y yum-utils ansible wget git net-tools bind-utils iptables-services bridge-utils bash-completion kexec-tools sos psacct docker openshift-ansible

# Disable spectre and meltdown patches
sed -i 's/quiet"/quiet spectre_v2=off nopti"/' /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

sed -i 's/--log-driver=journald //g' /etc/sysconfig/docker
echo '{ "insecure-registries" : ["registry:5000"] }' > /etc/docker/daemon.json

systemctl start docker
systemctl enable docker

dnsmasq_ip="192.168.66.2"
echo "$dnsmasq_ip nfs" >> /etc/hosts
echo "$dnsmasq_ip registry" >> /etc/hosts

# Allow connecting to ssh via password
sed -i -e "s/PasswordAuthentication no/PasswordAuthentication yes/" /etc/ssh/sshd_config
systemctl restart sshd

# Disable host key checking under ansible.cfg file
sed -i '/host_key_checking/s/^#//g' /etc/ansible/ansible.cfg

inventory_file="/root/inventory"
master_ip="192.168.66.101"
echo "$master_ip node01" >> /etc/hosts

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
openshift_image_tag=v3.9.0
ansible_service_broker_registry_whitelist=['.*-apb$']
openshift_hosted_etcd_storage_kind=nfs
openshift_hosted_etcd_storage_nfs_options="*(rw,root_squash,sync,no_wdelay)"
openshift_hosted_etcd_storage_nfs_directory=/opt/etcd-vol
openshift_hosted_etcd_storage_volume_name=etcd-vol
openshift_hosted_etcd_storage_access_modes=["ReadWriteOnce"]
openshift_hosted_etcd_storage_volume_size=1G
openshift_hosted_etcd_storage_labels={'storage': 'etcd'}
openshift_node_kubelet_args={'max-pods': ['40'], 'pods-per-core': ['40']}
openshift_master_admission_plugin_config={"ValidatingAdmissionWebhook":{"configuration":{"kind": "DefaultAdmissionConfig","apiVersion": "v1","disable": false}},"MutatingAdmissionWebhook":{"configuration":{"kind": "DefaultAdmissionConfig","apiVersion": "v1","disable": false}}}

[nfs]
node01 openshift_ip=$master_ip

[masters]
node01 openshift_ip=$master_ip

[etcd]
node01 openshift_ip=$master_ip

[nodes]
node01 openshift_node_labels="{'region': 'infra','zone': 'default'}" openshift_schedulable=true openshift_ip=$master_ip
EOF

# Add cri-o variable to inventory file
if [[ $1 == "true" ]]; then
    sed -i 's/\[OSEv3\:vars\]/\[OSEv3\:vars\]\nopenshift_use_crio=true\nopenshift_crio_systemcontainer_image_override=docker.io\/kubevirtci\/crio:1.9.10/' $inventory_file
fi

# Run OpenShift ansible playbook
ansible-playbook -e "ansible_user=root ansible_ssh_pass=vagrant" -i $inventory_file /usr/share/ansible/openshift-ansible/playbooks/prerequisites.yml
ansible-playbook -i $inventory_file /usr/share/ansible/openshift-ansible/playbooks/deploy_cluster.yml

# Create OpenShift user
/usr/bin/oc create user admin
/usr/bin/oc create identity allow_all_auth:admin
/usr/bin/oc create useridentitymapping allow_all_auth:admin admin
/usr/bin/oc adm policy add-cluster-role-to-user cluster-admin admin
