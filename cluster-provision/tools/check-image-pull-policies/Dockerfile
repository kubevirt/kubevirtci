FROM golang:alpine AS builder

WORKDIR /tmp/build
COPY . /tmp/build
RUN CGO_ENABLED=0 GOOS=linux GOPROXY=off GOFLAGS=-mod=vendor go build -a -o check-image-pull-policies ./check-image-pull-policies/.

FROM alpine:latest
COPY --from=builder /tmp/build/check-image-pull-policies /
ENTRYPOINT ["/check-image-pull-policies"]
