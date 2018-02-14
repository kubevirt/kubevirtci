FROM fedora:27

RUN dnf -y install docker && dnf clean all

COPY cli.sh /cli.sh

ENTRYPOINT [ "/cli.sh" ]

WORKDIR /
