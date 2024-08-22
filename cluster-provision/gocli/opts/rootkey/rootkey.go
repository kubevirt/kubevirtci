package rootkey

import (
	_ "embed"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type rootKey struct {
	sshClient libssh.Client
}

//go:embed conf/vagrant.pub
var key []byte

func NewRootKey(sc libssh.Client) *rootKey {
	return &rootKey{
		sshClient: sc,
	}
}

func (r *rootKey) Exec() error {
	cmds := []string{
		"echo '" + string(key) + "' | sudo tee /root/.ssh/authorized_keys > /dev/null",
		"sudo service sshd restart",
	}

	for _, cmd := range cmds {
		if err := r.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}
