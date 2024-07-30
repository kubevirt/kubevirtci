package istio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestIstioOpt(t *testing.T) {
	r := k8s.NewReactorConfig("create", "istiooperators", IstioReactor)
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	k8sclient := k8s.NewTestClient(r)

	opt := NewIstioOpt(sshClient, k8sclient, false)
	AddExpectCalls(sshClient)

	err := opt.Exec()
	assert.NoError(t, err)
}
