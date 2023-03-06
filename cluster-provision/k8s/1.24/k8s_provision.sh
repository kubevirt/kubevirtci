#!/bin/bash

set -ex

source /var/lib/kubevirtci/shared_vars.sh

cni_manifest="/provision/cni.yaml"

cp /tmp/local-volume.yaml /provision/local-volume.yaml

kubeadm_raw=/tmp/kubeadm.conf
if [[ ${KUBEVIRTCI_DUALSTACK} == false ]]; then
    kubeadm_raw=/tmp/kubeadm_ipv6.conf
fi

# TODO use config file! this is deprecated
cat <<EOT >/etc/sysconfig/kubelet
KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice  --fail-swap-on=false --kubelet-cgroups=/systemd/system.slice
EOT

# Needed for kubernetes service routing and dns
# https://github.com/kubernetes/kubernetes/issues/33798#issuecomment-250962627
modprobe bridge
modprobe overlay
modprobe br_netfilter
cat <<EOF >  /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
net.ipv6.conf.all.disable_ipv6 = 0
net.ipv6.conf.all.forwarding = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF
sysctl --system

echo bridge >> /etc/modules-load.d/k8s.conf
echo br_netfilter >> /etc/modules-load.d/k8s.conf
echo overlay >> /etc/modules-load.d/k8s.conf

# Delete conf files created by crio / podman
# so calico will create the interfaces by its own according the right configuration.
# See https://github.com/cri-o/cri-o/issues/2411#issuecomment-540006558
# It should happen before crio start, see https://github.com/cri-o/cri-o/issues/4276
# About podman see https://github.com/kubernetes/kubernetes/issues/107687
rm -f /etc/cni/net.d/*

systemctl daemon-reload
systemctl enable crio kubelet --now

dnf install -y NetworkManager NetworkManager-ovs

# configure additional settings for cni plugin
cat <<EOF >/etc/NetworkManager/conf.d/001-calico.conf
[keyfile]
unmanaged-devices=interface-name:cali*;interface-name:tunl*
EOF

# Use dhclient to have expected hostname behaviour
cat <<EOF >/etc/NetworkManager/conf.d/002-dhclient.conf
[main]
dhcp=dhclient
EOF

echo "net.netfilter.nf_conntrack_max=1000000" >> /etc/sysctl.conf
sysctl --system

systemctl restart NetworkManager

nmcli connection modify "System eth0" \
   ipv6.method auto \
   ipv6.addr-gen-mode eui64
nmcli connection up "System eth0"

kubeadmn_patches_path="/provision/kubeadm-patches"
mkdir -p $kubeadmn_patches_path

cat >$kubeadmn_patches_path/kube-apiserver.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF
cat >$kubeadmn_patches_path/kube-controller-manager.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF
cat >$kubeadmn_patches_path/kube-scheduler.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF
cat >$kubeadmn_patches_path/etcd.yaml <<EOF
spec:
  securityContext:
    seLinuxOptions:
      type: spc_t
EOF

cat >$kubeadmn_patches_path/add-security-context-deployment-patch.yaml <<EOF
spec:
  template:
    spec:
      securityContext:
        seLinuxOptions:
          type: spc_t
EOF

# audit log configuration
mkdir /etc/kubernetes/audit

audit_api_version="audit.k8s.io/v1"
cat > /etc/kubernetes/audit/adv-audit.yaml <<EOF
apiVersion: ${audit_api_version}
kind: Policy
rules:
- level: Request
  users: ["kubernetes-admin"]
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

#
cat > /etc/kubernetes/psa.yaml <<EOF
apiVersion: apiserver.config.k8s.io/v1
kind: AdmissionConfiguration
plugins:
- name: PodSecurity
  configuration:
    apiVersion: pod-security.admission.config.k8s.io/v1beta1
    kind: PodSecurityConfiguration
    defaults:
      enforce: "privileged"
      enforce-version: "latest"
      audit: "restricted"
      audit-version: "latest"
      warn: "restricted"
      warn-version: "latest"
    exemptions:
      usernames: []
      runtimeClasses: []
      # Hopefuly this will not be needed in future. Add your favorite namespace to be ignored and your operator not broken
      # You also need to modify psa.sh
      namespaces: ["kube-system", "default", "istio-operator" ,"istio-system", "nfs-csi", "monitoring", "rook-ceph", "cluster-network-addons", "sonobuoy"]
EOF

kubeadm_manifest="/etc/kubernetes/kubeadm.conf"
envsubst < $kubeadm_raw > $kubeadm_manifest

until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done


# prepull coredns:v1.8.6 and retag to expected tag
podman pull registry.k8s.io/coredns/coredns:v1.8.6 && podman tag registry.k8s.io/coredns/coredns:v1.8.6 k8s.gcr.io/coredns:v1.8.6

# 1.23 has deprecated --experimental-patches /provision/kubeadm-patches/, we now mention the patch directory in kubeadm.conf
kubeadm init --config $kubeadm_manifest -v5

kubectl --kubeconfig=/etc/kubernetes/admin.conf patch deployment coredns -n kube-system -p "$(cat $kubeadmn_patches_path/add-security-context-deployment-patch.yaml)"
kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f "$cni_manifest"

# Wait at least for 7 pods
while [[ "$(kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system --no-headers | wc -l)" -lt 7 ]]; do
    echo "Waiting for at least 7 pods to appear ..."
    kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system
    sleep 10
done

# Wait until k8s pods are running
while [ -n "$(kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system --no-headers | grep -v Running)" ]; do
    echo "Waiting for k8s pods to enter the Running state ..."
    kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system --no-headers | >&2 grep -v Running || true
    sleep 10
done

# Make sure all containers are ready
while [ -n "$(kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -o'custom-columns=status:status.containerStatuses[*].ready,metadata:metadata.name' --no-headers | grep false)" ]; do
    echo "Waiting for all containers to become ready ..."
    kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -o'custom-columns=status:status.containerStatuses[*].ready,metadata:metadata.name' --no-headers
    sleep 10
done

kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system

kubeadm reset --force

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

# copy network addons operator manifests
# so we can use them at cluster-up
cp -rf /tmp/cnao/ /opt/

# copy whereabouts manifests
# so we can use them at cluster-up
cp -rf /tmp/whereabouts/ /opt/

# copy Multus CNI manifests so we can use them at cluster-up
cp -rf /tmp/multus /opt/

# copy cdi manifests
cp -rf /tmp/cdi*.yaml /opt/

# Create a properly labelled tmp directory for testing
mkdir -p /var/provision/kubevirt.io/tests
chcon -t container_file_t /var/provision/kubevirt.io/tests
echo "tmpfs /var/provision/kubevirt.io/tests tmpfs rw,context=system_u:object_r:container_file_t:s0 0 1" >> /etc/fstab

dnf install -y NetworkManager-config-server

# Cleanup the existing NetworkManager profiles so the VM instances will come
# up with the default profiles. (Base VM image includes non default settings)
rm -f /etc/sysconfig/network-scripts/ifcfg-*
nmcli connection add con-name eth0 ifname eth0 type ethernet

# Remove machine-id, allowing unique ID/s for its instances
rm -f /etc/machine-id ; touch /etc/machine-id
