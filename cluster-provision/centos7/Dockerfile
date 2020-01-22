FROM kubevirtci/base@sha256:850ac2e2828610b5f35f004f2a8a1ab23609a4c7891c8a1b68cbb7eef5f5dda0

ARG centos_version

RUN echo "Centos7 version $centos_version"

ENV CENTOS_URL https://cloud.centos.org/centos/7/vagrant/x86_64/images/CentOS-7-x86_64-Vagrant-$centos_version.Libvirt.box

RUN /download_box.sh ${CENTOS_URL}
