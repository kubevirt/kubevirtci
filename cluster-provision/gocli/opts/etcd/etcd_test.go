package etcdinmemory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRealTimeOpt(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewEtcdInMemOpt(sshClient, 2020, 1, "512M")

	cmds := []string{
		"mkdir -p /var/lib/etcd",
		"test -d /var/lib/etcd",
		fmt.Sprintf("mount -t tmpfs -o size=%s tmpfs /var/lib/etcd", opt.etcdSize),
	}
	for _, cmd := range cmds {
		sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, cmd, true, true)
	}

	err := opt.Exec()
	assert.NoError(t, err)
}
