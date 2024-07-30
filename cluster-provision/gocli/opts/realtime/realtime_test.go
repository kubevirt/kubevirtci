package realtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRealTimeOpt(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewRealtimeOpt(sshClient)

	sshClient.EXPECT().Command("echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf", true)
	sshClient.EXPECT().Command("sysctl --system", true)
	err := opt.Exec()
	assert.NoError(t, err)
}
