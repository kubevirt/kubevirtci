FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

ARG TARGETARCH
ARG KUBEVIRTCI_TAG

WORKDIR /src
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build \
    -mod=vendor \
    -ldflags "-X 'kubevirt.io/kubevirtci/cluster-provision/gocli/images.SUFFIX=:${KUBEVIRTCI_TAG}'" \
    -o /cli ./cmd/cli

FROM alpine:3.23.2

RUN apk add git
WORKDIR /workdir
RUN git config --global --add safe.directory /workdir

COPY --from=builder /cli /cli

ENTRYPOINT ["/cli"]
