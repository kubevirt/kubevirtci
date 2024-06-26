package istio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestIstioOpt(t *testing.T) {
	updateIstioOperator := func(action k8stesting.Action) (bool, runtime.Object, error) {
		createAction := action.(k8stesting.CreateAction)
		obj := createAction.GetObject().(*unstructured.Unstructured)
		status := map[string]interface{}{
			"status": "HEALTHY",
		}
		if err := unstructured.SetNestedField(obj.Object, status, "status"); err != nil {
			return true, nil, err
		}
		return false, obj, nil
	}

	r := k8s.NewReactorConfig("create", "istiooperators", updateIstioOperator)
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	k8sclient := k8s.NewTestClient(r)

	opt := NewIstioOpt(sshClient, k8sclient, 2022, false)
	cmds := []string{
		"source /var/lib/kubevirtci/shared_vars.sh",
		"PATH=/opt/istio-" + opt.version + "/bin:$PATH istioctl --kubeconfig /etc/kubernetes/admin.conf --hub quay.io/kubevirtci operator init",
	}

	for _, cmd := range cmds {
		sshClient.EXPECT().JumpSSH(opt.sshPort, 1, cmd, true, true)
	}

	err := opt.Exec()
	assert.NoError(t, err)
}
