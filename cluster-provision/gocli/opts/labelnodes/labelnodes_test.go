package labelnodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestNodeLabel(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewNodeLabler(sshClient, "node-role.kubernetes.io/control-plane")
	AddExpectCalls(sshClient, opt.labelSelector)

	err := opt.Exec()
	assert.NoError(t, err)
}
