FROM alpine:3.23.2

RUN apk add git
WORKDIR /workdir
RUN git config --global --add safe.directory /workdir

COPY cli /cli

ENTRYPOINT ["/cli"]
