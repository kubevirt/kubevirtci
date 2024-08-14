package rootkey

import (
	"embed"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type RootKey struct {
	sshClient libssh.Client
}

//go:embed conf/*
var f embed.FS

func NewRootKey(sc libssh.Client) *RootKey {
	return &RootKey{
		sshClient: sc,
	}
}

func (r *RootKey) Exec() error {
	key, err := f.ReadFile("conf/vagrant.pub")
	if err != nil {
		return nil
	}

	cmds := []string{
		"echo '" + string(key) + "' | sudo tee /root/.ssh/authorized_keys > /dev/null",
		"sudo systemctl restart sshd",
	}

	for _, cmd := range cmds {
		if _, err := r.sshClient.Command(cmd, false); err != nil {
			return err
		}
	}

	return nil
}
