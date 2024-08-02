package labelnodes

import (
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type nodeLabler struct {
	sshClient     libssh.Client
	labelSelector string
}

func NewNodeLabler(sc libssh.Client, p uint16, l string) *nodeLabler {
	return &nodeLabler{
		sshClient:     sc,
		labelSelector: l,
	}
}

func (n *nodeLabler) Exec() error {
	if err := n.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf label node -l " + n.labelSelector + " node-role.kubernetes.io/worker=''"); err != nil {
		return err
	}
	return nil
}
