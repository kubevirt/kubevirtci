FROM fedora:27

RUN dnf -y install docker && dnf clean all

COPY cli /cli

ENTRYPOINT [ "/cli" ]

WORKDIR /
