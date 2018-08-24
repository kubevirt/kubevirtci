#!/bin/bash

set -ex

# Install epel
yum -y install epel-release

# Install storage requirements for iscsi and cluster
yum -y install centos-release-gluster
yum -y install --nogpgcheck -y glusterfs-fuse
yum -y install iscsi-initiator-utils

# Create Origin latest repo, enter correct repository address
cat >/etc/yum.repos.d/origin-latest.repo <<EOF
[my-origin]
name=Origin packages v3.10.0-rc.0
baseurl=https://plain.resources.ovirt.org/repos/origin/3.10/v3.10.0-rc.0/
enabled=1
gpgcheck=0
EOF

# Create Origin 3.10 testing repo
cat >/etc/yum.repos.d/origin-test.repo <<EOF
[centos-openshift-origin-test]
name=CentOS OpenShift Origin 3.10 test
baseurl=https://cbs.centos.org/repos/paas7-openshift-origin310-testing/x86_64/os/
enabled=1
gpgcheck=0
EOF

# Install OpenShift packages
yum install -y yum-utils \
  ansible \
  wget \
  git \
  net-tools \
  bind-utils \
  iptables-services \
  bridge-utils \
  bash-completion \
  kexec-tools \
  sos \
  psacct \
  docker

# Disable spectre and meltdown patches
sed -i 's/quiet"/quiet spectre_v2=off nopti hugepagesz=2M hugepages=64"/' /etc/default/grub
grub2-mkconfig -o /boot/grub2/grub.cfg

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

openshift_ansible="/root/openshift-ansible"
inventory_file="/root/inventory"
master_ip="192.168.66.101"
echo "$master_ip node01" >> /etc/hosts

git clone https://github.com/openshift/openshift-ansible.git -b v3.10.0-rc.0 $openshift_ansible

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
openshift_image_tag=v3.10.0-rc.0
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
os_sdn_network_plugin_name=redhat/openshift-ovs-networkpolicy

[nfs]
node01 openshift_ip=$master_ip

[masters]
node01 openshift_ip=$master_ip

[etcd]
node01 openshift_ip=$master_ip

[nodes]
node01 openshift_schedulable=true openshift_ip=$master_ip openshift_node_group_name="node-config-master-infra"
EOF

# Add cri-o variable to inventory file
if [[ $1 == "true" ]]; then
    sed -i 's/\[OSEv3\:vars\]/\[OSEv3\:vars\]\nopenshift_use_crio=true\n/' $inventory_file
fi

# Install prerequisites
ansible-playbook -e "ansible_user=root ansible_ssh_pass=vagrant" -i $inventory_file $openshift_ansible/playbooks/prerequisites.yml
ansible-playbook -i $inventory_file $openshift_ansible/playbooks/deploy_cluster.yml

# Create OpenShift user
/usr/bin/oc create user admin
/usr/bin/oc create identity allow_all_auth:admin
/usr/bin/oc create useridentitymapping allow_all_auth:admin admin
/usr/bin/oc adm policy add-cluster-role-to-user cluster-admin admin

# Create local-volume directories
for i in {1..10}
do
  mkdir -p /var/local/kubevirt-storage/local-volume/disk${i}
  mkdir -p /mnt/local-storage/local/disk${i}
  echo "/var/local/kubevirt-storage/local-volume/disk${i} /mnt/local-storage/local/disk${i} none defaults,bind 0 0" >> /etc/fstab
done
chmod -R 777 /var/local/kubevirt-storage/local-volume

# Setup selinux permissions to local volume directories.
chcon -R unconfined_u:object_r:svirt_sandbox_file_t:s0 /mnt/local-storage/
# Add privileged to local volume provision service account
/usr/bin/oc adm policy add-scc-to-user privileged -z local-storage-admin

