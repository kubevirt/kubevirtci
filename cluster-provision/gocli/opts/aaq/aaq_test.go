package aaq

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestAaqOpt(t *testing.T) {
	client := k8s.NewTestClient()
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))

	opt := NewAaqOpt(client, sshClient, "")
	sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait --for=condition=Ready pod --timeout=180s --all --namespace aaq", true)

	err := opt.Exec()
	assert.NoError(t, err)
}
