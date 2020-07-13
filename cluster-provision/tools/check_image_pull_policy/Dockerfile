FROM golang:alpine AS builder

WORKDIR /tmp/build
COPY . /tmp/build
RUN CGO_ENABLED=0 GOOS=linux GOPROXY=off GOFLAGS=-mod=vendor go build -a -o check_image_pull_policy ./check_image_pull_policy/.

FROM alpine:latest
COPY --from=builder /tmp/build/check_image_pull_policy /
ENTRYPOINT ["/check_image_pull_policy"]
