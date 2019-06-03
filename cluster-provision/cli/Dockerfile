#FROM fedora:27
FROM fedora@sha256:25f7dac76b2c88d8b7e0b1d6213d3406e77c7f230bfa1e66bd1cbb81a944eaaf

RUN dnf -y install docker && dnf clean all

COPY cli /cli

ENTRYPOINT [ "/cli" ]

WORKDIR /
