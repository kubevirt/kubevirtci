FROM kubevirtci/base@sha256:cb9a1cb7740f9b717d4988ffe3d5a7e41f38ed9b2fe9013458b225de19b7c3b4

ARG fedora_version

RUN echo "Fedora31 version $fedora_version"

ENV FEDORA_URL http://mirror.isoc.org.il/pub/fedora/releases/31/Cloud/x86_64/images/Fedora-Cloud-Base-Vagrant-$fedora_version.x86_64.vagrant-libvirt.box

RUN /download_box.sh ${FEDORA_URL}
