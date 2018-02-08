FROM fedora:27

RUN dnf -y update nettle && dnf -y install iptables iproute dnsmasq qemu openssh-clients && dnf clean all

WORKDIR /

COPY scripts/download_box.sh /download_box.sh

ENV CENTOS_VERSION 1801_02

RUN ./download_box.sh ${CENTOS_VERSION} && qemu-img convert -O qcow2 box.img box.qcow2 && rm box.img

COPY vagrant.key /vagrant.key

RUN chmod 700 vagrant.key

ENV DOCKERIZE_VERSION v0.3.0

RUN curl -LO https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && tar -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && chmod u+x dockerize \
  && mv dockerize /usr/local/bin/

COPY scripts/dnsmasq.sh /dnsmasq.sh

COPY scripts/provision.sh /provision.sh

COPY scripts/vm.sh /vm.sh
