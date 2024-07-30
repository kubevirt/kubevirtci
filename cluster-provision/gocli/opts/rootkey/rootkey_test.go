package rootkey

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRootKey(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewRootKey(sshClient)
	key, err := f.ReadFile("conf/vagrant.pub")

	sshClient.EXPECT().Command("echo '"+string(key)+"' | sudo tee /root/.ssh/authorized_keys > /dev/null", false)
	sshClient.EXPECT().Command("sudo service sshd restart", false)
	err = opt.Exec()
	assert.NoError(t, err)
}
