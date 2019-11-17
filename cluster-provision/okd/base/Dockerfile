FROM fedora@sha256:a66c6fa97957087176fede47846e503aeffc0441050dd7d6d2ed9e2fae50ea8e

RUN dnf install -y \
libvirt \
libvirt-devel \
libvirt-daemon-kvm \
libvirt-client \
qemu-kvm \
openssh-clients \
haproxy \
jq \
virt-install \
socat \
selinux-policy \
selinux-policy-targeted \
httpd-tools \
python2-pip && \
dnf clean all

RUN pip2 install yq

# install golang
ENV GO_VERSION 1.12.12
RUN curl https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz -o go.tar.gz && \
tar -xvzf go.tar.gz -C /usr/local/ && \
rm -rf go.tar.gz

# Install oc client
ENV OC_VERSION 4.3
RUN curl https://mirror.openshift.com/pub/openshift-v4/clients/oc/${OC_VERSION}/linux/oc.tar.gz -o oc.tar.gz && \
tar -xvzf oc.tar.gz -C /usr/local/bin/ && \
rm -rf oc.tar.gz

# configure libvirt
RUN echo 'listen_tls = 0' >> /etc/libvirt/libvirtd.conf; \
echo 'listen_tcp = 1' >> /etc/libvirt/libvirtd.conf; \
echo 'auth_tcp="none"' >> /etc/libvirt/libvirtd.conf; \
echo 'tcp_port = "16509"' >> /etc/libvirt/libvirtd.conf; \
echo 'cgroup_controllers = [ ]' >> /etc/libvirt/qemu.conf

COPY vagrant.key /
RUN chmod 600 /vagrant.key

COPY haproxy.cfg /etc/haproxy

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
