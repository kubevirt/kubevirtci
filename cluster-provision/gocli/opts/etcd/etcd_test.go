package etcdinmemory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRealTimeOpt(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewEtcdInMemOpt(sshClient, "512M")
	AddExpectCalls(sshClient, opt.etcdSize)

	err := opt.Exec()
	assert.NoError(t, err)
}
