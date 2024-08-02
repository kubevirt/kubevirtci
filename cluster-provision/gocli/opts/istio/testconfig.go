package istio

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

var IstioReactor = func(action k8stesting.Action) (bool, runtime.Object, error) {
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

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient) {
	cmds := []string{
		"source /var/lib/kubevirtci/shared_vars.sh",
		"PATH=/opt/istio-1.15.0/bin:$PATH istioctl --kubeconfig /etc/kubernetes/admin.conf --hub quay.io/kubevirtci operator init",
	}

	for _, cmd := range cmds {
		sshClient.EXPECT().Command(cmd)
	}
}
