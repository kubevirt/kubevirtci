package labelnodes

import (
	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type NodeLabler struct {
	sshClient     utils.SSHClient
	sshPort       uint16
	labelSelector string
}

func NewNodeLabler(sc utils.SSHClient, p uint16, l string) *NodeLabler {
	return &NodeLabler{
		sshClient:     sc,
		sshPort:       p,
		labelSelector: l,
	}
}

func (n *NodeLabler) Exec() error {
	if _, err := n.sshClient.JumpSSH(n.sshPort, 1, "kubectl --kubeconfig=/etc/kubernetes/admin.conf label node -l "+n.labelSelector+" node-role.kubernetes.io/worker=''", true, true); err != nil {
		return err
	}
	return nil
}
