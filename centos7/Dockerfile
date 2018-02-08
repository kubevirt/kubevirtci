FROM fedora:27

RUN dnf -y update nettle && dnf -y install iptables iproute dnsmasq qemu openssh-clients && dnf clean all

WORKDIR /

COPY download_box.sh /download_box.sh

ENV CENTOS_VERSION 1801_02

RUN ./download_box.sh ${CENTOS_VERSION} && qemu-img convert -O qcow2 box.img box.qcow2 && rm box.img

COPY vagrant.key /vagrant.key

RUN chmod 700 vagrant.key

COPY dnsmasq.sh /dnsmasq.sh

COPY vm.sh /vm.sh
