#!/bin/bash

set -xe

if [ -z $INSTALLER_COMMIT ]; then
    if [ ! -z $INSTALLER_RELEASE_IMAGE ]; then
        until  export INSTALLER_COMMIT=$(oc adm release info $INSTALLER_RELEASE_IMAGE --commits | grep installer | awk '{print $3}' | head -n 1); do
            sleep 1
        done
    fi
fi

compile_installer () {
    # install build dependencies
    local build_pkgs="git golang-bin gcc-c++"
    dnf install -y ${build_pkgs}

    # get installer code
    local installer_dir="/root/go/src/github.com/openshift/installer"
    mkdir -p ${installer_dir}
    cd ${installer_dir}
    git clone https://github.com/openshift/installer.git ${installer_dir}

    if [ ! -z $INSTALLER_COMMIT ]; then
        git checkout $INSTALLER_COMMIT
    else
        git checkout $INSTALLER_TAG
    fi

    # compile the installer
    if [ -d "/hacks" ] && [ ! -z $INSTALLER_TAG ]; then
        git apply /hacks/$INSTALLER_TAG
    fi

    if [ -d "/hacks" ] && [ ! -z $INSTALLER_COMMIT ]; then
        git apply /hacks/$INSTALLER_COMMIT
    fi

    TAGS=libvirt ./hack/build.sh
    cp bin/openshift-install /

    # clean after the compilation
    cd /
    rm -rf ${installer_dir}
    dnf erase -y ${build_pkgs} && dnf clean all
}

compile_installer

until virsh list
do
    sleep 5
done

# create libvirt storage pool
virsh pool-define /dev/stdin <<EOF
<pool type='dir'>
  <name>default</name>
  <target>
    <path>/var/lib/libvirt/images</path>
  </target>
</pool>
EOF
virsh pool-start default
virsh pool-autostart default

# dnsmasq configuration
original_dnss=$(cat /etc/resolv.conf | egrep "^nameserver" | awk '{print $2}')
echo "nameserver 127.0.0.1" > /etc/resolv.conf

mkdir -p /etc/dnsmasq.d
echo "server=/tt.testing/192.168.126.1" >> /etc/dnsmasq.d/openshift.conf
echo "address=/apps.tt.testing/192.168.126.51" >> /etc/dnsmasq.d/openshift.conf
for dns in $original_dnss; do
    echo "server=/#/$dns" >> /etc/dnsmasq.d/openshift.conf
done

/usr/sbin/dnsmasq \
--no-resolv \
--keep-in-foreground \
--no-hosts \
--bind-interfaces \
--pid-file=/var/run/dnsmasq.pid \
--listen-address=127.0.0.1 \
--cache-size=400 \
--clear-on-reload \
--conf-file=/dev/null \
--proxy-dnssec \
--strict-order \
--conf-file=/etc/dnsmasq.d/openshift.conf &

# wait until dnsmasq will start
sleep 5

if [ ! -f "/etc/installer/token" ]; then
    echo "You need to provide installer token file the container"
    exit 1
fi

mkdir -p /root/install
cp /install-config.yaml /root/install/

# inject pull secret into install config
cat /etc/installer/token >> /root/install/install-config.yaml

# inject vagrant ssh public key into install config
ssh_pub_key="ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzIw+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoPkcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NOTd0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcWyLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQ== vagrant insecure public key"
echo "sshKey: '$ssh_pub_key'" >> /root/install/install-config.yaml

if [ ! -z $INSTALLER_RELEASE_IMAGE ]; then
    export OPENSHIFT_INSTALL_RELEASE_IMAGE_OVERRIDE=$INSTALLER_RELEASE_IMAGE
fi

# increase workers memory to 6144MB
/openshift-install create manifests --dir=/root/install
sed -i -e "s/domainMemory: 4096/domainMemory: $WORKERS_MEMORY/" /root/install/openshift/99_openshift-cluster-api_worker-machineset-0.yaml
sed -i -e "s/domainVcpu: 2/domainVcpu: $WORKERS_CPU/" /root/install/openshift/99_openshift-cluster-api_worker-machineset-0.yaml

# run installer
export TF_VAR_libvirt_master_memory=$MASTER_MEMORY
export TF_VAR_libvirt_master_vcpu=$MASTER_CPU
/openshift-install create cluster --dir=/root/install --log-level=debug

export KUBECONFIG=/root/install/auth/kubeconfig

# Create OpenShift user
oc create user admin
oc create identity allow_all_auth:admin
oc create useridentitymapping allow_all_auth:admin admin
oc adm policy add-cluster-role-to-user cluster-admin admin

# Apply network addons
oc create -f /manifests/cna/namespace.yaml
oc create -f /manifests/cna/network-addons-config.crd.yaml
oc create -f /manifests/cna/operator.yaml
oc create -f /manifests/cna/network-addons-config-example.cr.yaml

 # Wait until all the network components are ready
oc wait networkaddonsconfig cluster --for condition=Ready --timeout=300s

# Make sure that all VMs can reach the internet
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i tt0 -o eth0 -j ACCEPT

# Shutdown VM's
virsh list --name | xargs --max-args=1 virsh shutdown

while [[ "$(virsh list --name)" != "" ]]; do
    sleep 1
done

# remove the cache
rm -rf /root/.cache/*
