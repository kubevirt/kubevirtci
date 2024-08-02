package etcdinmemory

import (
	"fmt"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type etcdInMemOpt struct {
	etcdSize  string
	sshClient libssh.Client
}

func NewEtcdInMemOpt(sc libssh.Client, size string) *etcdInMemOpt {
	if size == "" {
		size = "512M"
	}
	return &etcdInMemOpt{
		etcdSize:  size,
		sshClient: sc,
	}
}

func (o *etcdInMemOpt) Exec() error {
	cmds := []string{
		"mkdir -p /var/lib/etcd",
		"test -d /var/lib/etcd",
		fmt.Sprintf("mount -t tmpfs -o size=%s tmpfs /var/lib/etcd", o.etcdSize),
	}
	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}
