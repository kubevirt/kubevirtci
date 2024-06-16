package rootkey

import (
	"embed"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type RootKey struct {
	sshPort uint16
	nodeIdx int
}

//go:embed conf/*
var f embed.FS

func NewRootKey(p uint16, i int) *RootKey {
	return &RootKey{
		sshPort: p,
		nodeIdx: i,
	}
}

func (r *RootKey) Exec() error {
	key, err := f.ReadFile("conf/vagrant.pub")
	if err != nil {
		return nil
	}

	cmds := []string{
		"echo '" + string(key) + "' | sudo tee /root/.ssh/authorized_keys > /dev/null",
		"sudo service sshd restart",
	}

	for _, cmd := range cmds {
		if _, err := utils.JumpSSH(r.sshPort, r.nodeIdx, cmd, true, false); err != nil {
			return err
		}
	}

	return nil
}
