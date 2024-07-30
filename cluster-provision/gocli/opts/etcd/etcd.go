package etcdinmemory

import (
	"fmt"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type EtcdInMemOpt struct {
	etcdSize  string
	sshClient libssh.Client
}

func NewEtcdInMemOpt(sc libssh.Client, size string) *EtcdInMemOpt {
	if size == "" {
		size = "512M"
	}
	return &EtcdInMemOpt{
		etcdSize:  size,
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
		if _, err := o.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}

	return nil
}
