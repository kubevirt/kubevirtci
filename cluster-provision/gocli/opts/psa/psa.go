package psa

import (
	_ "embed"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/psa.yaml
var psa []byte

type psaOpt struct {
	sshClient libssh.Client
}

func NewPsaOpt(sc libssh.Client) *psaOpt {
	return &psaOpt{
		sshClient: sc,
	}
}

func (o *psaOpt) Exec() error {
	cmds := []string{
		"rm /etc/kubernetes/psa.yaml",
		"echo '" + string(psa) + "' | sudo tee /etc/kubernetes/psa.yaml > /dev/null",
	}
	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}
