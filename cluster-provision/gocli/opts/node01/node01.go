package node01

import (
	_ "embed"
	"fmt"
	"runtime"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed conf/00-cgroupv2.conf
var cgroupv2 []byte

//go:embed conf/adv-audit.yaml
var advAudit []byte

type node01Provisioner struct {
	sshClient   libssh.Client
	singleStack bool
	flannel     bool
	etcdNoFsync bool
}

func NewNode01Provisioner(sc libssh.Client, singleStack, flannel, etcdNoFsync bool) *node01Provisioner {
	return &node01Provisioner{
		sshClient:   sc,
		singleStack: singleStack,
		flannel:     flannel,
		etcdNoFsync: etcdNoFsync,
	}
}

func (n *node01Provisioner) Exec() error {
	var (
		kubeadmConf = "/etc/kubernetes/kubeadm.conf"
		cniManifest = "/provision/cni.yaml"
	)

	if n.flannel {
		kubeadmConf = "/etc/kubernetes/kubeadm_flannel.conf"
		cniManifest = "/etc/kubernetes/flannel.yaml"
	}

	if n.singleStack {
		if n.flannel {
			return fmt.Errorf("error: flannel single stack is not supported yet")
		}
		kubeadmConf = "/etc/kubernetes/kubeadm_ipv6.conf"
		cniManifest = "/provision/cni_ipv6.yaml"
	}

	kubeadmInitCmd := "kubeadm init --config " + kubeadmConf + " -v5"
	if n.etcdNoFsync {
		kubeadmInitCmd = fmt.Sprintf("sed -i 's/#etcdExtraArgs/extraArgs: \\{unsafe-no-fsync: \\\"True\\\"}/' %s && %s", kubeadmConf, kubeadmInitCmd)
	}

	cmds := []string{
		`if [ -f /home/` + libssh.GetUserByArchitecture(runtime.GOARCH) + `/enable_audit ]; then echo '` + string(advAudit) + `' | tee /etc/kubernetes/audit/adv-audit.yaml > /dev/null; fi`,
		`timeout=30; interval=5; while ! hostnamectl | grep Transient; do echo "Waiting for dhclient to set the hostname from dnsmasq"; sleep $interval; timeout=$((timeout - interval)); [ $timeout -le 0 ] && exit 1; done`,
		"swapoff -a",
		"until ip address show dev eth0 | grep global | grep inet6; do sleep 1; done",
		`timeout=60; interval=5; while ! systemctl status crio | grep -w "active"; do echo "Waiting for cri-o service to be ready"; sleep $interval; timeout=$((timeout - interval)); if [[ $timeout -le 0 ]]; then exit 1; fi; done`,
		kubeadmInitCmd,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf patch deployment coredns -n kube-system -p "$(cat /provision/kubeadm-patches/add-security-context-deployment-patch.yaml)"`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f ` + cniManifest,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf taint nodes node01 node-role.kubernetes.io/control-plane:NoSchedule-`,
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers; kubectl_rc=$?; retry_counter=0; while [[ $retry_counter -lt 20 && $kubectl_rc -ne 0 ]]; do sleep 10; echo "Waiting for api server to be available...";  kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --no-headers; kubectl_rc=$?; retry_counter=$((retry_counter + 1)); done`,
		"kubectl --kubeconfig=/etc/kubernetes/admin.conf version",
		`kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /provision/local-volume.yaml`,
		"mkdir -p /var/lib/rook",
		"chcon -t container_file_t /var/lib/rook",
	}
	for _, cmd := range cmds {
		err := n.sshClient.Command(cmd)
		if err != nil {
			return fmt.Errorf("error executing %s: %s", cmd, err)
		}
	}

	if n.flannel {
		cmd := `kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /etc/kubernetes/knp.yaml`
		err := n.sshClient.Command(cmd)
		if err != nil {
			return fmt.Errorf("error executing %s: %s", cmd, err)
		}
	}
	return nil
}
