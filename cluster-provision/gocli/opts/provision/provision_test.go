package provision

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestLinuxProvisioner(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewLinuxProvisioner(sshClient)
	AddExpectCalls(sshClient)

	err := opt.Exec()
	assert.NoError(t, err)
}
