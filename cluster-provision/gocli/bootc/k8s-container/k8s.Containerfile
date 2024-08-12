FROM LINUX_BASE

ARG VERSION

RUN echo -e "[isv_kubernetes_addons_cri-o_stable_v1.28]\n\
name=CRI-O v1.28 (Stable) (rpm)\n\
type=rpm-md\n\
baseurl=https://storage.googleapis.com/kubevirtci-crio-mirror/isv_kubernetes_addons_cri-o_stable_v1.28\n\
gpgcheck=0\n\
enabled=1" > /etc/yum.repos.d/devel_kubic_libcontainers_stable_cri-o_v1.28.repo

RUN MAJOR_MINOR=$(echo $VERSION | awk -F. '{print $1"."$2}') && \
    echo -e "[kubernetes]\n\
name=Kubernetes Release\n\
baseurl=https://pkgs.k8s.io/core:/stable:/v${MAJOR_MINOR}/rpm\n\
enabled=1\n\
gpgcheck=0\n\
repo_gpgcheck=0" > /etc/yum.repos.d/kubernetes.repo

RUN dnf install --nobest --nogpgcheck --disableexcludes=kubernetes -y \
    kubectl-${VERSION} \
    kubeadm-${VERSION} \
    kubelet-${VERSION} \
    kubernetes-cni


RUN dnf install -y cri-o patch

RUN mkdir -p /provision/kubeadm-patches

# COPY manifests /opt/
# COPY patches /provision/kubeadm-patches