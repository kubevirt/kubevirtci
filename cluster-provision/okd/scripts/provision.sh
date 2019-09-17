#!/bin/bash

set -xe

if [ ! -z $INSTALLER_RELEASE_IMAGE ]; then
    until  export INSTALLER_COMMIT=$(oc adm release info $INSTALLER_RELEASE_IMAGE --commits | grep installer | awk '{print $3}' | head -n 1); do
        sleep 1
    done
fi

compile_installer () {
    # install build dependencies
    local build_pkgs="git gcc-c++"
    dnf install -y ${build_pkgs}

    # install golang
    go_version=1.12.12
    curl https://dl.google.com/go/go${go_version}.linux-amd64.tar.gz -o go.tar.gz
    tar -xvzf go.tar.gz -C /usr/local/
    rm -rf go.tar.gz
    export GOROOT=/usr/local/go
    export GOPATH=/root/go/
    export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

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
    if [ -d "/hacks" ]; then
        git apply /hacks/$INSTALLER_TAG
    fi

    TAGS=libvirt ./hack/build.sh
    cp bin/openshift-install /

    # clean after the compilation
    cd /
    rm -rf ${installer_dir} ${GOROOT}
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
    echo "You need to provide installer token file to the container"
    exit 1
fi

export CLUSTER_DIR=/root/install
INSTALL_CONFIG_FILE=$CLUSTER_DIR/install-config.yaml

# we need to install jq and yq to make yaml changes
dnf install -y jq
pip install yq

function yq_inline {
    local expression="$1"
    local file="$2"
    if [ ! -f "$file" ]; then
        echo "$file is not a file!"
        return 1
    fi
    tmp_file=$(mktemp /tmp/output.XXXXXXXXXX)
    yq -y "$expression" "$file" > "$tmp_file"
    if [ $? -ne 0 ]; then
        return $?
    fi
    mv "$file" "$file.tmp"
    mv "$tmp_file" "$file"
}

mkdir -p $CLUSTER_DIR

# print version
/openshift-install version

cp /install-config.yaml $CLUSTER_DIR/

if [[ $INSTALLER_TAG =~ ^.*4\.[23]$ ]]; then

    # modify number of workers
    yq_inline '.compute[].replicas = 2' "$INSTALL_CONFIG_FILE"
fi

# inject pull secret into install config
cat /etc/installer/token >> $CLUSTER_DIR/install-config.yaml

# inject vagrant ssh public key into install config
ssh_pub_key="ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzIw+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoPkcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NOTd0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcWyLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQ== vagrant insecure public key"
echo "sshKey: '$ssh_pub_key'" >> $CLUSTER_DIR/install-config.yaml

# Generate manifests
/openshift-install create manifests --dir=$CLUSTER_DIR

# increase master memory
yq_inline '.spec.providerSpec.value.domainMemory = 8192' "$CLUSTER_DIR/openshift/99_openshift-cluster-api_master-machines-0.yaml"

# change workers memory and vcpu
yq_inline '.spec.template.spec.providerSpec.value.domainMemory = '"$WORKERS_MEMORY"' | .spec.template.spec.providerSpec.value.domainVcpu = '"$WORKERS_CPU" \
        $CLUSTER_DIR/openshift/99_openshift-cluster-api_worker-machineset-0.yaml

if [[ $INSTALLER_TAG =~ ^.*4\.[23]$ ]]; then

    # generate machineconfig for insecure-registries beforehand

    cat > registries.conf << __EOF__
[registries]
  [registries.search]
    registries = ["registry.access.redhat.com", "docker.io"]
  [registries.insecure]
    registries = ["brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888", "registry:5000"]
  [registries.block]
    registries = []
__EOF__

    cat > registries.yaml << __EOF__
spec:
  config:
    ignition:
      config: {}
      security:
        tls: {}
      timeouts: {}
      version: 2.2.0
    networkd: {}
    passwd: {}
    storage: {
            "files": [
                {
                    "path": "/etc/containers/registries.conf",
                    "filesystem": "root",
                    "mode": 420,
                    "contents": {
                    "source": "data:;base64,$(base64 -w0 registries.conf)"
                    }
                }
            ]
        }
    systemd: {
        "units": [
            {
                "contents": "[Unit]\nDescription=Update system CA\nAfter=syslog.target network.target\n\n[Service]\nType=oneshot\nExecStart=/usr/bin/update-ca-trust\nRemainAfterExit=true\n\n[Install]\nWantedBy=multi-user.target\n",
                "enabled": true,
                "name": "update-ca.service"
            }
        ]
    }
  osImageURL: ""
__EOF__

    cat > "${CLUSTER_DIR}/openshift/99-master-registries.yaml" << __EOF__
---
apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  labels:
    machineconfiguration.openshift.io/role: master
  name: 99-master-registries
$(cat registries.yaml)
__EOF__

    cat > "${CLUSTER_DIR}/openshift/99-worker-registries.yaml" << __EOF__
---
apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  labels:
    machineconfiguration.openshift.io/role: worker
  name: 99-worker-registries
$(cat registries.yaml)
__EOF__

    # for debug
    cp "${CLUSTER_DIR}/openshift/99-master-registries.yaml" ./
    cp "${CLUSTER_DIR}/openshift/99-worker-registries.yaml" ./

    # Generate ignition configs
    /openshift-install --dir "${CLUSTER_DIR}" create ignition-configs

fi

if [ ! -z $INSTALLER_RELEASE_IMAGE ]; then
    export OPENSHIFT_INSTALL_RELEASE_IMAGE_OVERRIDE=$INSTALLER_RELEASE_IMAGE
fi

# Excecute installer
export TF_VAR_libvirt_master_memory=$MASTER_MEMORY
export TF_VAR_libvirt_master_vcpu=$MASTER_CPU
/openshift-install create cluster --dir "$CLUSTER_DIR" --log-level debug


export KUBECONFIG=$CLUSTER_DIR/auth/kubeconfig

# Create htpasswd with user admin
htpasswd -c -B -b /root/htpasswd admin admin

# Create OpenShift HTPasswd provider with user and password admin
oc create secret generic htpass-secret --from-file=htpasswd=/root/htpasswd -n openshift-config
oc apply -f - <<EOF
apiVersion: config.openshift.io/v1
kind: OAuth
metadata:
  name: cluster
spec:
  identityProviders:
  - name: htpasswd_provider
    mappingMethod: claim
    type: HTPasswd
    htpasswd:
      fileData:
        name: htpass-secret
EOF

# Grant to admin user cluster-admin permissions
oc adm policy add-cluster-role-to-user cluster-admin admin

# Apply network addons
oc create -f /manifests/cna/namespace.yaml
oc create -f /manifests/cna/network-addons-config.crd.yaml
oc create -f /manifests/cna/operator.yaml
oc create -f /manifests/cna/network-addons-config-example.cr.yaml

 # Wait until all the network components are ready
oc wait networkaddonsconfig cluster --for condition=Ready --timeout=600s

# Remove master schedulable taint from masters
masters=$(oc get nodes -l node-role.kubernetes.io/master -o'custom-columns=name:metadata.name' --no-headers)
for master in ${masters}; do
    oc adm taint nodes ${master} node-role.kubernetes.io/master-
done

if [[ $INSTALLER_TAG =~ ^.*4\.1$ ]]; then
  # Add registry:5000 to insecure registries
  until oc patch image.config.openshift.io/cluster --type merge --patch '{"spec": {"registrySources": {"insecureRegistries": ["registry:5000", "brew-pulp-docker01.web.prod.ext.phx2.redhat.com:8888"]}}}'; do
      sleep 5
  done

  until [[ $(oc get nodes --no-headers | grep master | grep Ready,SchedulingDisabled | wc -l) -ge 1 ]]; do
      sleep 10
  done

  # Make master nodes schedulable
  for master in ${masters}; do
      oc adm uncordon ${master}
  done
fi

until [[ $(oc get nodes --no-headers | grep -v SchedulingDisabled | grep Ready | wc -l) -ge 2 ]]; do
    sleep 10
done

# Enable CPU manager on workers
until oc label machineconfigpool worker custom-kubelet=cpumanager-enabled; do
    sleep 5
done

oc create -f - <<EOF
apiVersion: machineconfiguration.openshift.io/v1
kind: KubeletConfig
metadata:
  name: cpumanager-enabled
spec:
  machineConfigPoolSelector:
    matchLabels:
      custom-kubelet: cpumanager-enabled
  kubeletConfig:
     cpuManagerPolicy: static
     cpuManagerReconcilePeriod: 5s
EOF

until [[ $(oc get nodes --no-headers | grep worker | grep SchedulingDisabled | wc -l) -ge 1 ]]; do
    sleep 10
done

until [[ $(oc get nodes --no-headers | grep -v SchedulingDisabled | grep Ready | wc -l) -ge 2 ]]; do
    sleep 10
done

# Disable updates of machines configurations, because on the update the machine-config
# controller will try to drain the master node, but it not possible with only one master
# so the node will stay in cordon state forewer.
# It will prevent any updates of the kubelet or registries configuration on the machine
# so if you need some, please add it before these lines

# Scale down cluster version operator to prevent updates of other operators
until oc -n openshift-cluster-version scale --replicas=0 deploy cluster-version-operator; do
    sleep 5
done

# Scale down machine-config-operator to prevent re-creation of master machineconfigpools
until oc -n openshift-machine-config-operator scale --replicas=0 deploy machine-config-operator; do
    sleep 5
done

# Delete machine-config-daemon to prevent configuration updates
until oc -n openshift-machine-config-operator delete ds machine-config-daemon; do
    sleep 5
done

# Create local storage objects under the cluster
oc create ns local-storage

oc create -f /manifests/okd/local-storage.yaml
until oc -n local-storage get LocalVolume; do
    sleep 5
done

if [[ $INSTALLER_TAG =~ ^.*4\.3$ ]]; then
    # workaround for 4.3 bz https://bugzilla.redhat.com/show_bug.cgi?id=1766856
    # TODO: Remove this when fix is in place at OCP release
    oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:local-storage:local-storage-operator
fi

oc create -f /manifests/okd/local-storage-cr.yaml
until oc -n local-storage get sc local; do
    sleep 5
done

# Set the default storage class
oc patch storageclass local -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

# Make sure that all VMs can reach the internet
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -A FORWARD -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i tt0 -o eth0 -j ACCEPT

# Shutdown VM's
virsh list --name | xargs --max-args=1 virsh shutdown

while [[ "$(virsh list --name)" != "" ]]; do
    sleep 1
done

# Remove the cache
rm -rf /root/.cache/*
