package labelnodes

import (
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type NodeLabler struct {
	sshClient     libssh.Client
	labelSelector string
}

func NewNodeLabler(sc libssh.Client, l string) *NodeLabler {
	return &NodeLabler{
		sshClient:     sc,
		labelSelector: l,
	}
}

func (n *NodeLabler) Exec() error {
	if _, err := n.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf label node -l "+n.labelSelector+" node-role.kubernetes.io/worker=''", true); err != nil {
		return err
	}
	return nil
}
