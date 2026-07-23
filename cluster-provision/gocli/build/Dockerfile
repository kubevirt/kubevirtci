FROM alpine:3.23.2

ARG TARGETARCH=amd64

RUN apk add git
WORKDIR /workdir
RUN git config --global --add safe.directory /workdir

COPY cli-${TARGETARCH} /cli

ENTRYPOINT ["/cli"]
