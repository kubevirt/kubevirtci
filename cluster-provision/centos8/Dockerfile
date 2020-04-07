
#FROM fedora:31
FROM fedora@sha256:8fa60b88e2a7eac8460b9c0104b877f1aa0cea7fbc03c701b7e545dacccfb433

ARG centos_version

RUN dnf -y update nettle && dnf -y install jq iptables iproute dnsmasq qemu openssh-clients socat && dnf clean all

WORKDIR /

COPY vagrant.key /vagrant.key

RUN chmod 700 vagrant.key

ENV DOCKERIZE_VERSION v0.3.0

RUN curl -LO https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && tar -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && chmod u+x dockerize \
  && mv dockerize /usr/local/bin/

COPY scripts/* /
RUN echo "Centos8 version $centos_version"

ENV CENTOS_URL https://cloud.centos.org/centos/8-stream/x86_64/images/CentOS-Stream-Vagrant-8-$centos_version.x86_64.vagrant-libvirt.box

RUN /download_box.sh ${CENTOS_URL}
