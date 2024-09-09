package rookceph

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

var CephReactor = func(action k8stesting.Action) (bool, runtime.Object, error) {
	createAction := action.(k8stesting.CreateAction)
	obj := createAction.GetObject().(*unstructured.Unstructured)
	status := map[string]interface{}{
		"phase": "Ready",
	}
	if err := unstructured.SetNestedField(obj.Object, status, "status"); err != nil {
		return true, nil, err
	}
	return false, obj, nil
}

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient) {
	sshClient.EXPECT().Command(`kubectl --kubeconfig /etc/kubernetes/admin.conf patch storageclass local -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'`)
	sshClient.EXPECT().Command(`kubectl --kubeconfig /etc/kubernetes/admin.conf patch storageclass rook-ceph-block -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'`)
}
