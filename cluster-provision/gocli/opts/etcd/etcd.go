package etcdinmemory

import (
	"fmt"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type EtcdInMemOpt struct {
	sshPort  uint16
	nodeIdx  int
	etcdSize string
}

func NewEtcdInMemOpt(p uint16, idx int, s string) *EtcdInMemOpt {
	if s == "" {
		s = "512M"
	}
	return &EtcdInMemOpt{
		sshPort:  p,
		nodeIdx:  idx,
		etcdSize: s,
	}
}

func (o *EtcdInMemOpt) Exec() error {
	cmds := []string{
		"sudo mkdir -p /var/lib/etcd",
		"sudo test -d /var/lib/etcd",
		fmt.Sprintf("sudo mount -t tmpfs -o size=%s tmpfs /var/lib/etcd", o.etcdSize),
	}
	for _, cmd := range cmds {
		if _, err := utils.JumpSSH(o.sshPort, o.nodeIdx, cmd, true, true); err != nil {
			return err
		}
	}

	return nil
}
