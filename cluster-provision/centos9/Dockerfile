# Build the vmcli Go CLI
FROM docker.io/library/golang:1.22 AS vmcli-builder

WORKDIR /app

COPY vmcli/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./_bin/vmcli ./cmd/cli

# Prepare the VM runner
FROM quay.io/kubevirtci/fedora@sha256:e3a6087f62f288571db14defb7e0e10ad7fe6f973f567b0488d3aac5e927035a AS runner

ARG centos_version

RUN dnf -y install jq iptables iproute dnsmasq qemu openssh-clients screen bind-utils tcpdump iputils && dnf clean all

WORKDIR /

COPY vagrant.key /vagrant.key

RUN chmod 700 vagrant.key

ENV DOCKERIZE_VERSION v0.6.1

RUN curl -LO https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && tar -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && rm dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
  && chmod u+x dockerize \
  && mv dockerize /usr/local/bin/

COPY scripts/download_box.sh /

RUN echo "Centos9 version $centos_version"

ENV CENTOS_URL https://cloud.centos.org/centos/9-stream/x86_64/images/CentOS-Stream-Vagrant-9-$centos_version.x86_64.vagrant-libvirt.box

RUN /download_box.sh ${CENTOS_URL}

RUN curl -L -o /initrd.img http://mirror.stream.centos.org/9-stream/BaseOS/x86_64/os/images/pxeboot/initrd.img
RUN curl -L -o /vmlinuz http://mirror.stream.centos.org/9-stream/BaseOS/x86_64/os/images/pxeboot/vmlinuz

COPY scripts/* /
COPY --from=vmcli-builder /app/_bin/vmcli /
