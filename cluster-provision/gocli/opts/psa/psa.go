package psa

import (
	"embed"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/*
var f embed.FS

type PsaOpt struct {
	sshClient libssh.Client
}

func NewPsaOpt(sc libssh.Client) *PsaOpt {
	return &PsaOpt{
		sshClient: sc,
	}
}

func (o *PsaOpt) Exec() error {
	psa, err := f.ReadFile("manifests/psa.yaml")
	if err != nil {
		return err
	}
	cmds := []string{
		"rm /etc/kubernetes/psa.yaml",
		"echo '" + string(psa) + "' | sudo tee /etc/kubernetes/psa.yaml > /dev/null",
	}
	for _, cmd := range cmds {
		if _, err := o.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}

	return nil
}
