FROM quay.io/fedora/fedora:39 AS base

RUN dnf -y install jq iptables iproute dnsmasq qemu socat openssh-clients screen bind-utils tcpdump iputils libguestfs-tools-c && dnf clean all


FROM base AS imageartifactdownload

ARG BUILDARCH

ARG centos_version

WORKDIR /

RUN echo "Centos9 version $centos_version"

COPY scripts/download_box.sh /

RUN if test "$BUILDARCH" != "s390x"; then \
      /download_box.sh https://cloud.centos.org/centos/9-stream/x86_64/images/CentOS-Stream-Vagrant-9-$centos_version.x86_64.vagrant-libvirt.box && \
      curl -L -o /initramfs-amd64.img http://mirror.stream.centos.org/9-stream/BaseOS/x86_64/os/images/pxeboot/initrd.img && \
      curl -L -o /vmlinuz-amd64 http://mirror.stream.centos.org/9-stream/BaseOS/x86_64/os/images/pxeboot/vmlinuz; \
    else \
      /download_box.sh https://cloud.centos.org/centos/9-stream/s390x/images/CentOS-Stream-GenericCloud-9-$centos_version.s390x.qcow2 && \
      # Access virtual machine disk images directly by using LIBGUESTFS_BACKEND=direct, instead of libvirt
      export LIBGUESTFS_BACKEND=direct && \
      guestfish --ro --add box.qcow2 --mount /dev/sda1:/ ls /boot/ | grep -E '^vmlinuz-|^initramfs-' | xargs -I {} guestfish --ro --add box.qcow2 -i copy-out /boot/{} / ; \
    fi


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