package ksm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestKsmOPt(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewKsmOpt(sshClient, 20, 10)

	cmds := []string{
		"echo 1 | sudo tee /sys/kernel/mm/ksm/run >/dev/null",
		"echo " + fmt.Sprintf("%d", opt.scanInterval) + " | sudo tee /sys/kernel/mm/ksm/sleep_millisecs >/dev/null",
		"echo " + fmt.Sprintf("%d", opt.pagesToScan) + " | sudo tee /sys/kernel/mm/ksm/pages_to_scan >/dev/null",
	}

	for _, cmd := range cmds {
		sshClient.EXPECT().Command(cmd, true)
	}
	err := opt.Exec()
	assert.NoError(t, err)
}
