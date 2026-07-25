FROM quay.io/centos/centos:stream9 AS base

RUN dnf -y install epel-release

RUN dnf -y install \
    bind-utils \
    dnsmasq \
    guestfs-tools \
    jq \
    iproute \
    iptables \
    iputils \
    openssh-clients \
    screen \
    socat \
    tcpdump \
    qemu-kvm-core && \
    dnf clean all

FROM base AS imageartifactdownload

ARG BUILDARCH=x86_64
ARG CENTOS_VERSION

WORKDIR /

RUN echo "CentOS Stream 9 version $CENTOS_VERSION" && \
    curl --fail-with-body --retry 2 -LR -o box.qcow2 https://cloud.centos.org/centos/9-stream/$BUILDARCH/images/CentOS-Stream-GenericCloud-9-$CENTOS_VERSION.$BUILDARCH.qcow2 && \
    # Access virtual machine disk images directly by using LIBGUESTFS_BACKEND=direct, instead of libvirt
    export LIBGUESTFS_BACKEND=direct && \
    guestfish --ro --add box.qcow2 --mount /dev/sda1:/ ls /boot/ | grep -E '^vmlinuz-|^initramfs-' | xargs -I {} guestfish --ro --add box.qcow2 -i copy-out /boot/{} / ;

FROM base AS nodecontainer

ARG BUILDARCH

WORKDIR /

COPY vagrant.key /vagrant.key

RUN chmod 700 vagrant.key

ENV DOCKERIZE_VERSION=v0.8.0

RUN if test "$BUILDARCH" != "s390x"; then \
      curl -L -o dockerize-linux-$BUILDARCH.tar.gz https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz; \
    else \
      curl -L -o dockerize-linux-$BUILDARCH.tar.gz https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-s390x-$DOCKERIZE_VERSION.tar.gz; \
    fi && \
    tar -xzvf dockerize-linux-$BUILDARCH.tar.gz && \
    rm dockerize-linux-$BUILDARCH.tar.gz && \
    chmod u+x dockerize && \
    mv dockerize /usr/local/bin/

COPY --from=imageartifactdownload /box.qcow2 box.qcow2
COPY --from=imageartifactdownload /vmlinuz-* /vmlinuz
COPY --from=imageartifactdownload /initramfs-* /initrd.img

COPY scripts/* /
