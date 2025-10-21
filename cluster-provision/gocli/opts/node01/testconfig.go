package node01

import (
	"fmt"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient) {
	cmds := []string{
		fmt.Sprintf(`if [ -f /home/%s/enable_audit ]; then echo '%s' | tee /etc/kubernetes/audit/adv-audit.yaml > /dev/null; fi`, libssh.GetSSHUser(), string(advAudit)),
		`timeout=30; interval=5; while ! hostnamectl | grep Transient; do echo "Waiting for dhclient to set the hostname from dnsmasq"; sleep $interval; timeout=$((timeout - interval)); [ $timeout -le 0 ] && exit 1; done`,
		"swapoff -a",
		`until PRIMARY_IFACE=$(ip -o addr show | awk '/192\.168\.66\./ {print $2; exit}'); [ -n "$PRIMARY_IFACE" ] && ip address show dev $PRIMARY_IFACE | grep global | grep inet6; do sleep 1; done`,
		`timeout=60; interval=5; while ! systemctl status crio | grep -w "active"; do echo "Waiting for cri-o service to be ready"; sleep $interval; timeout=$((timeout - interval)); if [[ $timeout -le 0 ]]; then exit 1; fi; done`,
		`kubeadm init --config /etc/kubernetes/kubeadm.conf -v5`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf patch deployment coredns -n kube-system -p "$(cat /provision/kubeadm-patches/add-security-context-deployment-patch.yaml)"`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /provision/cni.yaml`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/control-plane:NoSchedule-`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers; kubectl_rc=$?; retry_counter=0; while [[ $retry_counter -lt 20 && $kubectl_rc -ne 0 ]]; do sleep 10; echo "Waiting for api server to be available...";  kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers; kubectl_rc=$?; retry_counter=$((retry_counter + 1)); done`,
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf version",
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /provision/local-volume.yaml`,
		"mkdir -p /var/lib/rook",
		"chcon -t container_file_t /var/lib/rook",
	}
	for _, cmd := range cmds {
		sshClient.EXPECT().Command(cmd)
	}
}
