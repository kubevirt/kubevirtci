package rootkey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRootKey(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewRootKey(sshClient, 2020, 1)
	key, err := f.ReadFile("conf/vagrant.pub")

	sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, "echo '"+string(key)+"' | sudo tee /root/.ssh/authorized_keys > /dev/null", false, false)
	sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, "sudo service sshd restart", false, false)
	err = opt.Exec()
	assert.NoError(t, err)
}
