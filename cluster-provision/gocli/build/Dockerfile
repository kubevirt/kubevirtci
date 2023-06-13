FROM fedora:35

RUN dnf install -y git
WORKDIR /workdir
RUN git config --global --add safe.directory /workdir

COPY cli /cli

ENTRYPOINT ["/cli"]
