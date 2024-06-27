package etcdinmemory

import (
	"fmt"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type EtcdInMemOpt struct {
	sshPort   uint16
	nodeIdx   int
	etcdSize  string
	sshClient utils.SSHClient
}

func NewEtcdInMemOpt(sc utils.SSHClient, p uint16, idx int, s string) *EtcdInMemOpt {
	if s == "" {
		s = "512M"
	}
	return &EtcdInMemOpt{
		sshPort:   p,
		nodeIdx:   idx,
		etcdSize:  s,
		sshClient: sc,
	}
}

func (o *EtcdInMemOpt) Exec() error {
	cmds := []string{
		"mkdir -p /var/lib/etcd",
		"test -d /var/lib/etcd",
		fmt.Sprintf("mount -t tmpfs -o size=%s tmpfs /var/lib/etcd", o.etcdSize),
	}
	for _, cmd := range cmds {
		if _, err := o.sshClient.JumpSSH(o.sshPort, o.nodeIdx, cmd, true, true); err != nil {
			return err
		}
	}

	return nil
}
