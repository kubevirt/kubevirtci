#!/bin/bash

set -xe

if [ ! -f "/etc/installer/token" ]; then
    echo "You need to provide installer pull secret file to the container"
    exit 1
fi

if [ ! -z $INSTALLER_RELEASE_IMAGE ]; then
    until  export INSTALLER_COMMIT=$(oc adm release info -a /etc/installer/token $INSTALLER_RELEASE_IMAGE --commits | grep installer | awk '{print $3}' | head -n 1); do
        sleep 1
    done
fi

compile_installer () {
    # install build dependencies
    local build_pkgs="git gcc-c++"
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
    if [ -d "/hacks" ]; then
        git apply /hacks/$INSTALLER_TAG
    fi

    GOROOT=/usr/local/go
    GOPATH=/root/go/
    PATH=$GOPATH/bin:$GOROOT/bin:$PATH
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
sleep 10

export CLUSTER_DIR=/root/install
INSTALL_CONFIG_FILE=$CLUSTER_DIR/install-config.yaml

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

# fill registries.yaml with registries from the conf
export REGISTRIES_CONF=$(base64 -w0 /manifests/okd/registries.conf)
envsubst < /manifests/okd/registries.yaml > /registries.yaml

# inject PULL_SECRET and SSH_PUBLIC_KEY into install-config
set +x
export PULL_SECRET=$(cat /etc/installer/token)
export SSH_PUBLIC_KEY="ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEA6NF8iallvQVp22WDkTkyrtvp9eWW6A8YVr+kz4TjGYe7gHzIw+niNltGEFHzD8+v1I2YJ6oXevct1YeS0o9HZyN1Q9qgCgzUFtdOKLv6IedplqoPkcmF0aYet2PkEDo3MlTBckFXPITAMzF8dJSIFo9D8HfdOV0IAdx4O7PtixWKn5y2hMNG0zQPyUecp4pzC6kivAIhyfHilFR61RGL+GPXQ2MWZWFYbAGjyiYJnAmCP3NOTd0jMZEnDkbUvxhMmBYSdETk1rRgm+R4LOzFUGaHqHDLKLX+FIPKcF96hrucXzcWyLbIbEgE98OHlnVYCzRdK8jlqm8tehUc9c9WhQ== vagrant insecure public key"
envsubst < /manifests/okd/install-config.yaml > ${INSTALL_CONFIG_FILE}
unset PULL_SECRET
set -x

if [ ! -z $INSTALLER_RELEASE_IMAGE ]; then
    export OPENSHIFT_INSTALL_RELEASE_IMAGE_OVERRIDE=$INSTALLER_RELEASE_IMAGE
fi

# Generate manifests
/openshift-install create manifests --dir=$CLUSTER_DIR

# change master memory and vcpu
yq_inline '.spec.providerSpec.value.domainMemory = '"$MASTER_MEMORY"' | .spec.providerSpec.value.domainVcpu = '"$MASTER_CPU" \
       $CLUSTER_DIR/openshift/99_openshift-cluster-api_master-machines-0.yaml

# change workers memory and vcpu
yq_inline '.spec.template.spec.providerSpec.value.domainMemory = '"$WORKERS_MEMORY"' | .spec.template.spec.providerSpec.value.domainVcpu = '"$WORKERS_CPU" \
        $CLUSTER_DIR/openshift/99_openshift-cluster-api_worker-machineset-0.yaml

cat > "${CLUSTER_DIR}/openshift/99-master-registries.yaml" << __EOF__
---
apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  labels:
    machineconfiguration.openshift.io/role: master
  name: 99-master-registries
$(cat /registries.yaml)
__EOF__

cat > "${CLUSTER_DIR}/openshift/99-worker-registries.yaml" << __EOF__
---
apiVersion: machineconfiguration.openshift.io/v1
kind: MachineConfig
metadata:
  labels:
    machineconfiguration.openshift.io/role: worker
  name: 99-worker-registries
$(cat /registries.yaml)
__EOF__

# for debug
cp "${CLUSTER_DIR}/openshift/99-master-registries.yaml" ./
cp "${CLUSTER_DIR}/openshift/99-worker-registries.yaml" ./

# Generate ignition configs
/openshift-install --dir "${CLUSTER_DIR}" create ignition-configs

# Excecute installer
export TF_VAR_libvirt_master_memory=$MASTER_MEMORY
export TF_VAR_libvirt_master_vcpu=$MASTER_CPU
/openshift-install create cluster --dir "$CLUSTER_DIR" --log-level debug

export KUBECONFIG=$CLUSTER_DIR/auth/kubeconfig

oc wait --for=condition=Ready $(oc get node -o name) --timeout=600s

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

if [ "${CNAO}" == "true" ]; then
    # Apply network addons
    oc create -f /manifests/cna/namespace.yaml
    oc create -f /manifests/cna/network-addons-config.crd.yaml
    oc create -f /manifests/cna/operator.yaml
    oc create -f /manifests/cna/network-addons-config-example.cr.yaml

     # Wait until all the network components are ready
    until oc wait networkaddonsconfig cluster --for condition=Available --timeout=100s; do
        sleep 10
    done
fi

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

oc -n openshift-machine-config-operator wait machineconfigpools worker --for condition=Updating --timeout=1800s
oc -n openshift-machine-config-operator wait machineconfigpools worker --for condition=Updated --timeout=1800s

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

# Remove master schedulable taint from masters
masters=$(oc get nodes -l node-role.kubernetes.io/master -o'custom-columns=name:metadata.name' --no-headers)
for master in ${masters}; do
    oc adm taint nodes ${master} node-role.kubernetes.io/master-
done

until [[ $(oc get nodes --no-headers | grep -v SchedulingDisabled | grep Ready | wc -l) -ge 3 ]]; do
    sleep 10
done

# Create local storage objects under the cluster
oc create ns local-storage

oc create -f /manifests/okd/local-storage.yaml
until oc -n local-storage get LocalVolume; do
    sleep 5
done

# Remove the pull-secret
until oc -n openshift-config patch secret pull-secret --type merge --patch '{"data": {".dockerconfigjson": "e30K"}}'; do
    sleep 5
done

# workaround for 4.3 bz https://bugzilla.redhat.com/show_bug.cgi?id=1766856
# TODO: Remove this when fix is in place at OCP release
oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:local-storage:local-storage-operator

oc create -f /manifests/okd/local-storage-cr.yaml
until oc -n local-storage get sc local; do
    sleep 5
done

# Set the default storage class
until oc patch storageclass local -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'; do
    sleep 5
done

# update number of workers
worker_machine_set=$(oc -n openshift-machine-api get machineset --no-headers | grep worker | awk '{print $1}')
until oc -n openshift-machine-api scale --replicas=1 machineset ${worker_machine_set}; do
    sleep 5
done

while [[ "$(oc get node -o name |wc -l)" -ne 2 ]]; do
    sleep 15
done

# Clean Completed and OOMKilled pods
for pod in $(oc get pod --all-namespaces -o 'jsonpath={range .items[*]}{.metadata.namespace}{'\'','\''}{.metadata.name}{'\'','\''}{.status.phase}{'\''\n'\''}{end}' --field-selector status.phase!=Running |grep -v Pending); do
    oc delete pod $(echo $pod |sed -r "s/^(.*),(.*),.*$/-n \1 \2/g")
done


# Shutdown VM's
virsh list --name | xargs --max-args=1 virsh shutdown

while [[ "$(virsh list --name)" != "" ]]; do
    sleep 1
done

# Remove the cache
rm -rf /root/.cache/*
