package multus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestMultusOpt(t *testing.T) {
	client := k8s.NewTestClient()
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))

	opt := NewMultusOpt(client, sshClient)
	sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf rollout status -n kube-system ds/kube-multus-ds --timeout=200s", true)

	err := opt.Exec()
	assert.NoError(t, err)
}
