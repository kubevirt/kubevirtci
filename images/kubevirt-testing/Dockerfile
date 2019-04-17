FROM fedora:28

LABEL maintainer="The KubeVirt Project <kubevirt-dev@googlegroups.com>"
ENV container docker

RUN dnf install -y \
        nginx \
        scsi-target-utils \
        procps-ng \
        nmap-ncat \
        e2fsprogs \
    && dnf -y clean all
