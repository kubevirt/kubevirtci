package bindvfio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestBindVfio(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewBindVfioOpt(sshClient, "8086:2668")
	AddExpectCalls(sshClient, opt.pciID)
	err := opt.Exec()
	assert.NoError(t, err)
}
