package psa

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRealTimeOpt(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewPsaOpt(sshClient, 2020)
	psa, _ := f.ReadFile("manifests/psa.yaml")

	sshClient.EXPECT().JumpSSH(opt.sshPort, 1, "rm /etc/kubernetes/psa.yaml", true, true)
	sshClient.EXPECT().JumpSSH(opt.sshPort, 1, "echo '"+string(psa)+"' | sudo tee /etc/kubernetes/psa.yaml > /dev/null", true, true)
	err := opt.Exec()
	assert.NoError(t, err)
}
