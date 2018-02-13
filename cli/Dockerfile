FROM fedora:27

RUN dnf -y install docker && dnf clean all

COPY run.sh /run.sh

ENTRYPOINT [ "/run.sh" ]

WORKDIR /
