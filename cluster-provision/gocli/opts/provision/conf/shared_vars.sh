set -ex
export KUBELET_CGROUP_ARGS="--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"
export ISTIO_VERSION=1.15.0
export ISTIO_BIN_DIR="/opt/istio-1.15.0/bin"