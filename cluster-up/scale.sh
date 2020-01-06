#!/usr/bin/bash -x

# Add Custom Worker

CUSTOM_IMAGE=${CUSTOM_IMAGE:-"0"}
PROVISION_MODE=${PROVISION_MODE:-"0"}
REPO_FILE=${REPO_FILE:-"0"}

export KUBECONFIG=/root/install/auth/kubeconfig

SELF_PID=$(echo $$)
ps -ef | grep scale.sh | grep -v grep | awk '{print $2}' | grep -v $SELF_PID | xargs kill
pkill dnf
pkill ansible-playbook

if [ ! -f /root/.ssh/id_rsa.pub ]
then
   echo "id_rsa tuple must exists, Generating"
   ssh-keygen -t rsa -f /root/.ssh/id_rsa -N ''
fi

dnf install -y libguestfs-tools
dnf install -y https://repo.fedora.md/centos/7/virt/x86_64/ovirt-4.2/common/ansible-2.8.2-1.el7.noarch.rpm
dnf install -y openshift-ansible openshift-clients jq

NETLIST=$(virsh net-list | grep -v def | grep active | awk '{print $1}')

cd /tmp
if [ $CUSTOM_IMAGE = "0" ] || [ $PROVISION_MODE = "1" ]; then
   qemu-img create -f qcow2 custom-worker-1.img 20G
   virt-resize --expand /dev/sda1 base.img custom-worker-1.img
else
   cp $CUSTOM_IMAGE custom-worker-1.img
fi

virt-customize -a custom-worker-1.img --root-password password:123456 --ssh-inject root:file:/root/.ssh/id_rsa.pub --selinux-relabel --timezone Europe/Berlin --hostname $NETLIST-worker-1 --uninstall cloud-init,kexec-tools,postfix
virt-install --name $NETLIST-worker-1 --description "$NETLIST-worker-1" --os-type=linux --os-variant=rhel7.6 --vcpus=4 --ram 7168 --rng /dev/urandom --import --disk custom-worker-1.img --network network=$NETLIST --check all=off --connect qemu:///system --graphics none -m 52:54:00:72:5c:30 --noautoconsole --cpu host-passthrough

function getip {
  IP=$(virsh domifaddr $NETLIST-worker-1  | awk '{ print $4 }' | tail -2 | head -1 | awk -F/ '{print $1}')
}

function run_ssh {
  ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -q -lroot -T -i ~/.ssh/id_rsa $IP $@
}

while true
do
   getip
   [ -z $IP ] || break
   sleep 5
done
echo $IP

if [ -f /etc/yum.repos.d/$REPO_FILE ]; then
  while ! scp -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no /etc/yum.repos.d/$REPO_FILE root@$IP:/etc/yum.repos.d; do
     sleep 5
  done
fi

# Enable nested virtualization
run_ssh rmmod kvm-intel
run_ssh "echo 'options kvm-intel nested=1' >> /etc/modprobe.d/dist.conf"
run_ssh modprobe kvm-intel

# Add vagrant public key
vagrant_pub_key=$(ssh-keygen -y -f /vagrant.key)
run_ssh "echo 'ssh-rsa $vagrant_pub_key vagrant insecure public key' >> /root/.ssh/authorized_keys"

# Add core user
CMDS=$(cat <<__EOF__
useradd core || true;
echo core | passwd --stdin core;
mkdir -p /home/core/.ssh;
cp .ssh/authorized_keys /home/core/.ssh;
chown -R core /home/core/.ssh;
__EOF__
)

run_ssh "$CMDS"

cp $KUBECONFIG /tmp
master_ip="192.168.126.11"
sed -i "s/127.0.0.1/$master_ip/" /tmp/kubeconfig

cat <<EOF > /usr/share/ansible/openshift-ansible/inventory/hosts
[all:vars]
ansible_user=root
openshift_kubeconfig_path="/tmp/kubeconfig"

[new_workers]
$IP
EOF


cd /usr/share/ansible/openshift-ansible
if [ $PROVISION_MODE = "1" ]; then
  match=' - name: Apply ignition manifest'
  insert=' - name: Goodbye\n    fail:\n      msg: All good\n    when: true\n'
  file='./roles/openshift_node/tasks/config.yml'
  sed -i.bak "s/$match/$insert\n $match/" $file

  ansible-playbook -i inventory/hosts playbooks/scaleup.yml

  # Shutdown worker
  DOMAIN_NAME=$(virsh net-list | grep -v def | grep active | awk '{print $1}')-worker-1

  function get_state {
    STATE=$(virsh dominfo $DOMAIN_NAME | grep State | awk '{print $2}')
  }

  run_ssh "rm -f /etc/yum.repos.d/$REPO_FILE"
  virsh shutdown $DOMAIN_NAME

  while true
  do
     get_state
     [ $STATE != "shut" ] || break
     sleep 5
  done

  mv /tmp/custom-worker-1.img /tmp/custom.img
else
  if [ $CUSTOM_IMAGE != "0" ]; then
    match='- name: Install openshift support packages'
    insert='  tags:\n    - skipme'
    file='./roles/openshift_node/tasks/install.yml'
    sed -i.bak "s/$match/$match\n$insert/" $file

    match='  - name: Install openshift packages'
    insert='    tags:\n      - skipme'
    file='./roles/openshift_node/tasks/install.yml'
    sed -i "s/$match/$match\n$insert/" $file

    match=' - name: Pull release image'
    insert='    tags:\n      - skipme'
    file='./roles/openshift_node/tasks/config.yml'
    sed -i.bak "s/$match/$match\n$insert/" $file

    match='  - name: Pull MCD image'
    insert='    tags:\n      - skipme'
    file='./roles/openshift_node/tasks/config.yml'
    sed -i "s/$match/$match\n$insert/" $file
  fi

  ansible-playbook -i inventory/hosts playbooks/scaleup.yml --skip-tags=skipme
fi

rm -rf /etc/yum.repos.d/$REPO_FILE
