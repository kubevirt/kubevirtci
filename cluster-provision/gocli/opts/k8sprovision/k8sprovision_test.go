package k8sprovision

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestK8sProvision(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewK8sProvisioner(sshClient, "1.30", true)
	AddExpectCalls(sshClient, opt.version, opt.slim)

	err := opt.Exec()
	assert.NoError(t, err)
}
