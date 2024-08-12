FROM quay.io/centos-bootc/centos-bootc:stream9

ENV KUBEVIRTCI_SHARED_DIR=/var/lib/kubevirtci
ENV ISTIO_VERSION=1.15.0
ENV ISTIO_BIN_DIR=/opt/istio-${ISTIO_VERSION}/bin

RUN dnf update -y

RUN mkdir -p /opt/scripts

COPY provision-system.sh /opt/scripts/provision-system.sh
RUN chmod 755 /opt/scripts/provision-system.sh
COPY provision-system.service /etc/systemd/system/provision-system.service

RUN echo "vagrant ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

RUN mkdir -p $KUBEVIRTCI_SHARED_DIR \
    && echo '#!/bin/bash\n' \
         'set -ex\n' \
         'export KUBELET_CGROUP_ARGS="--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"\n' \
         'export ISTIO_VERSION=${ISTIO_VERSION}\n' \
         'export ISTIO_BIN_DIR="/opt/istio-${ISTIO_VERSION}/bin"\n' \
         > $KUBEVIRTCI_SHARED_DIR/shared_vars.sh \
    && chmod +x $KUBEVIRTCI_SHARED_DIR/shared_vars.sh

RUN dnf install -y "kernel-modules-5.14.0-480.el9.x86_64" \
    && dnf install -y patch \
    && dnf install -y pciutils \
    && systemctl enable provision-system.service \
    && dnf -y remove firewalld \
    && dnf -y install iscsi-initiator-utils \
    && dnf -y install nftables \
    && dnf -y install lvm2 \
    && echo 'ACTION=="add|change", SUBSYSTEM=="block", KERNEL=="vd[a-z]", ATTR{queue/rotational}="0"' > /etc/udev/rules.d/60-force-ssd-rotational.rules \
    && dnf install -y iproute-tc \
    && mkdir -p "$ISTIO_BIN_DIR" \
    && curl "https://storage.googleapis.com/kubevirtci-istioctl-mirror/istio-${ISTIO_VERSION}/bin/istioctl" -o "$ISTIO_BIN_DIR/istioctl" \
    && chmod +x "$ISTIO_BIN_DIR/istioctl" \
    && dnf install -y container-selinux \
    && dnf install -y libseccomp-devel \
    && dnf install -y centos-release-nfv-openvswitch \
    && dnf install -y openvswitch2.16 \
    && dnf install -y --skip-broken NetworkManager NetworkManager-ovs NetworkManager-config-server \
    && dnf clean all \
    && rm -rf /lib/systemd/system/systemd-zram-setup@.service

ENV PATH="$ISTIO_BIN_DIR:$PATH"