FROM fedora@sha256:a66c6fa97957087176fede47846e503aeffc0441050dd7d6d2ed9e2fae50ea8e

# oc binary without SSL dependency
ENV OC_PKG https://artifacts-openshift-master.svc.ci.openshift.org/repo/openshift-clients-4.0.0-0.alpha.0.1969.af45cda.x86_64.rpm

RUN dnf install -y \
${OC_PKG} \
libvirt \
libvirt-devel \
libvirt-daemon-kvm \
libvirt-client \
qemu-kvm \
openssh-clients \
haproxy \
virt-install \
socat \
selinux-policy \
selinux-policy-targeted \
httpd-tools && \
dnf clean all

# configure libvirt
RUN echo 'listen_tls = 0' >> /etc/libvirt/libvirtd.conf; \
echo 'listen_tcp = 1' >> /etc/libvirt/libvirtd.conf; \
echo 'auth_tcp="none"' >> /etc/libvirt/libvirtd.conf; \
echo 'tcp_port = "16509"' >> /etc/libvirt/libvirtd.conf; \
echo 'cgroup_controllers = [ ]' >> /etc/libvirt/qemu.conf

COPY vagrant.key /
RUN chmod 600 /vagrant.key

COPY haproxy.cfg /etc/haproxy
COPY install-config.yaml /

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT [ "/entrypoint.sh" ]
