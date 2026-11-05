package nodes

import (
	"go.uber.org/mock/gomock"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/common"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient) {
	cmds := []string{
		"source /var/lib/kubevirtci/shared_vars.sh",
		common.WaitForHostnameCommand("node02"),
		`echo "KUBELET_EXTRA_ARGS=--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice --fail-swap-on=false  --feature-gates=NodeSwap=true --cpu-manager-policy=static --kube-reserved=cpu=500m --system-reserved=cpu=500m" | tee /etc/sysconfig/kubelet > /dev/null`,
		"systemctl daemon-reload &&  service kubelet restart",
		"swapoff -a",
		"until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done",
		`timeout=60; interval=5; while ! systemctl status crio | grep -w "active"; do echo "Waiting for cri-o service to be ready"; sleep $interval; timeout=$((timeout - interval)); if [[ $timeout -le 0 ]]; then exit 1; fi; done`,
		"kubeadm join --token abcdef.1234567890123456 192.168.66.101:6443 --ignore-preflight-errors=all --discovery-token-unsafe-skip-ca-verification=true",
		"mkdir -p /var/lib/rook",
		"chcon -t container_file_t /var/lib/rook",
	}

	calls := make([]any, 0, len(cmds))
	for _, cmd := range cmds {
		calls = append(calls, sshClient.EXPECT().Command(cmd))
	}
	gomock.InOrder(calls...)
}
