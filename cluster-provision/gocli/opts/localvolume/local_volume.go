package localvolume

import (
	_ "embed"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/local-volume.yaml
var lv []byte

type localVolumeOpt struct {
	sshClient libssh.Client
}

func NewLocalVolumeOpt(sshClient libssh.Client) *localVolumeOpt {
	return &localVolumeOpt{
		sshClient: sshClient,
	}
}

func (lv *localVolumeOpt) Exec() error {
	if err := lv.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf create -f /provision/local-volume.yaml"); err != nil {
		return err
	}
	return nil
}
