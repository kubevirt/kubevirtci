package cnao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestCnaoOpt(t *testing.T) {
	client := k8s.NewTestClient()
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))

	opt := NewCnaoOpt(client, sshClient)
	sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s", true)

	err := opt.Exec()
	assert.NoError(t, err)
}
