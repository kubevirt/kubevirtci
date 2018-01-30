FROM fedora:27

RUN dnf -y install iproute dnsmasq qemu && dnf clean all

COPY centos7_vagrant_box_image_0.img /image.qcow2

RUN dnf -y update nettle && dnf -y install iptables openssh-clients && dnf clean all

COPY vagrant.key /vagrant.key

COPY dnsmasq.sh /dnsmasq.sh

COPY vm.sh /vm.sh
