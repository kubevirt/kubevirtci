#!/bin/bash

set -ex

# Install epel
yum -y install epel-release

# Install storage requirements for iscsi and cluster
yum -y install centos-release-gluster
yum -y install --nogpgcheck -y glusterfs-fuse
yum -y install iscsi-initiator-utils

# Create Origin latest repo, enter correct repository address
cat >/etc/yum.repos.d/origin-311.repo <<EOF
[origin-311]
name=Origin packages v3.11.0
baseurl=http://mirror.centos.org/centos/7/paas/x86_64/openshift-origin311/
enabled=1
gpgcheck=0
EOF

cat >/etc/yum.repos.d/ansible.repo <<EOF
[Ansible]
name=Ansible
baseurl=https://releases.ansible.com/ansible/rpm/release/epel-7-x86_64/
enabled=1
gpgcheck=0
EOF


# Install OpenShift packages
yum install -y ansible-2.7.11-1.el7.ans \
  wget \
  git \
  net-tools \
  bind-utils \
  yum-utils \
  iptables-services \
  bridge-utils \
  bash-completion \
  kexec-tools \
  sos \
  psacct \
  docker-common-1.13.1-75.git8633870.el7.centos.x86_64 \
  origin-docker-excluder-3.11.0-1.el7.git.0.62803d0.noarch \
  python-docker-py-1.10.6-4.el7.noarch \
  docker-client-1.13.1-75.git8633870.el7.centos.x86_64 \
  cockpit-docker-176-2.el7.centos.x86_64 \
  docker-1.13.1-75.git8633870.el7.centos.x86_64 \
  python-docker-pycreds-1.10.6-4.el7.noarch

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

git clone https://github.com/openshift/openshift-ansible.git -b v3.11.0 --depth 1 $openshift_ansible

# Apply fix https://github.com/openshift/openshift-ansible/pull/10459
# TODO: remove it when the fix will be available under the v3.11.0 tag
sed -i 's/python-docker/python-docker-py/' $openshift_ansible/playbooks/init/base_packages.yml

# Create ansible inventory file
cat >$inventory_file <<EOF
all:
  children:
    OSEv3:
      hosts:
        node01:
          openshift_ip: $master_ip
          openshift_node_group_name: node-config-master-infra-kubevirt
          openshift_schedulable: true
      children:
        masters:
          hosts:
            node01:
        nodes:
          hosts:
            node01:
        nfs:
          hosts:
            node01:
        etcd:
          hosts:
            node01:
      vars:
        openshift_enable_service_catalog: false
        ansible_ssh_pass: vagrant
        ansible_ssh_user: root
        deployment_type: origin
        openshift_clock_enabled: true
        openshift_deployment_type: origin
        openshift_disable_check: memory_availability,disk_availability,docker_storage,package_availability,docker_image_availability
        openshift_hosted_etcd_storage_access_modes:
        - ReadWriteOnce
        openshift_hosted_etcd_storage_kind: nfs
        openshift_hosted_etcd_storage_labels:
          storage: etcd
        openshift_hosted_etcd_storage_nfs_directory: /opt/etcd-vol
        openshift_hosted_etcd_storage_nfs_options: '*(rw,root_squash,sync,no_wdelay)'
        openshift_hosted_etcd_storage_volume_name: etcd-vol
        openshift_hosted_etcd_storage_volume_size: 1G
        openshift_image_tag: v3.11.0
        os_sdn_network_plugin_name: redhat/openshift-ovs-networkpolicy
        openshift_master_admission_plugin_config:
          MutatingAdmissionWebhook:
            configuration:
              apiVersion: v1
              disable: false
              kind: DefaultAdmissionConfig
          ValidatingAdmissionWebhook:
            configuration:
              apiVersion: v1
              disable: false
              kind: DefaultAdmissionConfig
        openshift_master_identity_providers:
        - challenge: 'true'
          kind: AllowAllPasswordIdentityProvider
          login: 'true'
          name: allow_all_auth
        osm_api_server_args:
          feature-gates:
          - BlockVolume=true
        osm_controller_args:
          feature-gates:
          - BlockVolume=true
        openshift_master_audit_config:
          enabled: true
          logFormat: json
          auditFilePath: "/var/lib/origin/audit-ocp.log"
          policyFile: "/etc/origin/master/adv-audit.yaml"
        openshift_node_groups:
        - name: node-config-master-infra-kubevirt
          labels:
          - node-role.kubernetes.io/master=true
          - node-role.kubernetes.io/infra=true
          - node-role.kubernetes.io/compute=true
          edits:
          - key: kubeletArguments.feature-gates
            value:
            - RotateKubeletClientCertificate=true,RotateKubeletServerCertificate=true,BlockVolume=true
          - key: kubeletArguments.max-pods
            value:
            - '60'
          - key: kubeletArguments.pods-per-core
            value:
            - '60'
        - name: node-config-compute-kubevirt
          labels:
          - node-role.kubernetes.io/compute=true
          edits:
          - key: kubeletArguments.feature-gates
            value:
            - RotateKubeletClientCertificate=true,RotateKubeletServerCertificate=true,BlockVolume=true,CPUManager=true
          - key: kubeletArguments.cpu-manager-policy
            value:
            - static
          - key: kubeletArguments.system-reserved
            value:
            - cpu=500m
          - key: kubeletArguments.kube-reserved
            value:
            - cpu=500m
          - key: kubeletArguments.max-pods
            value:
            - '60'
          - key: kubeletArguments.pods-per-core
            value:
            - '60'
EOF

mkdir -p /etc/origin/master
cat >/etc/origin/master/adv-audit.yaml <<EOF
apiVersion: audit.k8s.io/v1beta1
kind: Policy
rules:
- level: Request
  users: ["system:admin"]
  resources:
  - group: kubevirt.io
    resources:
    - virtualmachines
    - virtualmachineinstances
    - virtualmachineinstancereplicasets
    - virtualmachineinstancepresets
    - virtualmachineinstancemigrations
  omitStages:
  - RequestReceived
  - ResponseStarted
  - Panic
EOF

# Add cri-o variable to inventory file
if [[ $1 == "true" ]]; then
    sed -i "s/    vars\:/    vars\:\n        openshift_use_crio: 'true'/" $inventory_file
fi

# Install prerequisites
ansible-playbook -e "ansible_user=root ansible_ssh_pass=vagrant" -i $inventory_file $openshift_ansible/playbooks/prerequisites.yml
ansible-playbook -i $inventory_file $openshift_ansible/playbooks/deploy_cluster.yml

# install the container networking cni plugins
wget https://github.com/containernetworking/plugins/releases/download/v0.8.0/cni-plugins-linux-amd64-v0.8.0.tgz -P /opt/cni/bin/
tar -xvf /opt/cni/bin/cni-plugins-linux-amd64-v0.8.0.tgz -C /opt/cni/bin/
rm -rf /opt/cni/bin/cni-plugins-linux-amd64-v0.8.0.tgz

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

# Pre pull fluentd image used in logging
docker pull docker.io/fluent/fluentd:v1.2-debian
docker pull fluent/fluentd-kubernetes-daemonset:v1.2-debian-syslog

/usr/bin/oc create -f /tmp/cna/namespace.yaml
/usr/bin/oc create -f /tmp/cna/network-addons-config.crd.yaml
/usr/bin/oc create -f /tmp/cna/operator.yaml

# Wait until flannel cluster-network-addons-operator and core dns pods are running
# Sleep to give the controller time to create the operator pod
sleep 10
while [ -n "$(/usr/bin/oc get pods --all-namespaces --no-headers | grep -v Running)" ]; do
    echo "Waiting for flannel cluster-network-addons-operator and core dns pods to enter the Running state ..."
    /usr/bin/oc get pods --all-namespaces --no-headers | >&2 grep -v Running || true
    sleep 10
done

/usr/bin/oc create -f /tmp/cna/network-addons-config-example.cr.yaml

# Wait until all the network components are ready
/usr/bin/oc wait networkaddonsconfig cluster --for condition=Ready --timeout=800s

/usr/bin/oc create -f /tmp/openshift-ovs-cni.yaml

# Wait before checking the pod status.
# Give time to the scheduler to create the pods.
sleep 10

while [[ $(/usr/bin/oc get po -n kube-system --no-headers | grep -v Running | wc -l) -ne 0 ]]
do
    echo "Waiting for all the containers to be running"
    sleep 5
done

# Wait for the daemonset to create the ovs file
while  [ ! -f /usr/bin/ovs-vsctl ]
do
    echo "ovs-vsctl not found!"
    sleep 5
done

# Create a bridge for the tests
ovs-vsctl add-br br1
