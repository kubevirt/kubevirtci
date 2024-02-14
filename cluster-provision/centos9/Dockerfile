
FROM quay.io/kubevirtci/fedora@sha256:e3a6087f62f288571db14defb7e0e10ad7fe6f973f567b0488d3aac5e927035a

ARG centos_version

RUN dnf -y install jq iptables iproute dnsmasq qemu openssh-clients screen bind-utils tcpdump iputils && dnf clean all

WORKDIR /

COPY vagrant.key /vagrant.key

RUN chmod 700 vagrant.key

ENV DOCKERIZE_VERSION v0.6.1

RUN curl -LO https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && tar -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && chmod u+x dockerize \
  && mv dockerize /usr/local/bin/

COPY scripts/download_box.sh /

RUN echo "Centos9 version $centos_version"

ENV CENTOS_URL https://cloud.centos.org/centos/9-stream/x86_64/images/CentOS-Stream-Vagrant-9-$centos_version.x86_64.vagrant-libvirt.box

RUN /download_box.sh ${CENTOS_URL}

# Pin to earlier versions of vmlinuz & initrd.img due to issue with latest
RUN curl -L -o /initrd.img https://composes.stream.centos.org/production/CentOS-Stream-9-20240126.0/compose/BaseOS/x86_64/os/images/pxeboot/initrd.img
RUN curl -L -o /vmlinuz https://composes.stream.centos.org/production/CentOS-Stream-9-20240126.0/compose/BaseOS/x86_64/os/images/pxeboot/vmlinuz

#RUN curl -L -o /initrd.img http://mirror.stream.centos.org/9-stream/BaseOS/x86_64/os/images/pxeboot/initrd.img
#RUN curl -L -o /vmlinuz http://mirror.stream.centos.org/9-stream/BaseOS/x86_64/os/images/pxeboot/vmlinuz

COPY scripts/* /
