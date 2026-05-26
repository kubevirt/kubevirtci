package vsock

import (
	"fmt"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type vsockOpt struct {
	sshClient libssh.Client
	mode      string
}

func NewVsockOpt(sc libssh.Client, mode string) (*vsockOpt, error) {
	if mode != "global" && mode != "local" {
		return nil, fmt.Errorf("invalid vsock child namespace mode %q, must be one of: global, local", mode)
	}
	return &vsockOpt{
		sshClient: sc,
		mode:      mode,
	}, nil
}

func (o *vsockOpt) Exec() error {
	cmds := []string{
		"modprobe vsock",
		fmt.Sprintf("sysctl --write net.vsock.child_ns_mode=%s", o.mode),
	}

	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}
	return nil
}
