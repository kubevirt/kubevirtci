FROM kubevirtci/base@sha256:5e9185e680a1fb1bbdb4a2ff5a23d2c8229fa9c6359d461e460ffc06d35d5f02

ARG centos_version

RUN echo "Centos8 version $centos_version"

ENV CENTOS_URL https://cloud.centos.org/centos/8-stream/x86_64/images/CentOS-Stream-Vagrant-8-$centos_version.x86_64.vagrant-libvirt.box

RUN /download_box.sh ${CENTOS_URL}
