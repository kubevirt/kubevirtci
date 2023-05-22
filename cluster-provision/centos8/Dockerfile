
FROM quay.io/kubevirtci/fedora@sha256:9b8e2b468d0d0523529bde44e8a0f17aad93dcedcf876c111183ffd4ea40d724

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

RUN echo "Centos8 version $centos_version"

ENV CENTOS_URL https://kojihub.stream.centos.org/kojifiles/packages/CentOS-Stream-Vagrant/8/$centos_version/images/CentOS-Stream-Vagrant-8-$centos_version.x86_64.vagrant-libvirt.box

RUN /download_box.sh ${CENTOS_URL}

RUN curl -L -o /initrd.img http://mirror.centos.org/centos/8-stream/BaseOS/x86_64/os/images/pxeboot/initrd.img
RUN curl -L -o /vmlinuz http://mirror.centos.org/centos/8-stream/BaseOS/x86_64/os/images/pxeboot/vmlinuz

COPY scripts/* /
