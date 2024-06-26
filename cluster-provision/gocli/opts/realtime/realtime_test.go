package realtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRealTimeOpt(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewRealtimeOpt(sshClient, 2020, 1)

	sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, "echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf", true, true)
	sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, "sysctl --system", true, true)
	err := opt.Exec()
	assert.NoError(t, err)
}
