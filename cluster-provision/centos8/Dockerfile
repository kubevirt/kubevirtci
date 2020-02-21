FROM kubevirtci/base@sha256:cb9a1cb7740f9b717d4988ffe3d5a7e41f38ed9b2fe9013458b225de19b7c3b4

ARG centos_version

RUN echo "Centos8 version $centos_version"

ENV CENTOS_URL https://cloud.centos.org/centos/8-stream/x86_64/images/CentOS-Stream-Vagrant-8-$centos_version.x86_64.vagrant-libvirt.box

RUN /download_box.sh ${CENTOS_URL}
