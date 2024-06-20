package rookceph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stesting "k8s.io/client-go/testing"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestWithFakeClient(t *testing.T) {
	updateBlockPool := func(action k8stesting.Action) (bool, runtime.Object, error) {
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

	r := k8s.NewReactorConfig("create", "cephblockpools", updateBlockPool)
	testClient := k8s.NewTestClient(r)
	opt := NewCephOpt(testClient)
	err := opt.Exec()
	assert.NoError(t, err)
}

func TestCephOpt(t *testing.T) {
	mockK8sClient := kubevirtcimocks.NewMockK8sDynamicClient(gomock.NewController(t))
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ceph.rook.io/v1",
			"kind":       "CephBlockPool",
			"metadata": map[string]interface{}{
				"name":      "replicapool",
				"namespace": "rook-ceph",
			},
			"status": map[string]interface{}{
				"phase": "Ready",
			},
			"spec": map[string]interface{}{
				"replicated": map[string]interface{}{
					"size": 1,
				},
			},
		},
	}

	opt := NewCephOpt(mockK8sClient)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/snapshot.storage.k8s.io_volumesnapshots.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/snapshot.storage.k8s.io_volumesnapshotcontents.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/snapshot.storage.k8s.io_volumesnapshotclasses.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/rbac-snapshot-controller.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/setup-snapshot-controller.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/common.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/crds.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/operator.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/cluster-test.yaml").Return(nil)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/pool-test.yaml").Return(nil)
	mockK8sClient.EXPECT().Get(schema.GroupVersionKind{
		Group:   "ceph.rook.io",
		Version: "v1",
		Kind:    "CephBlockPool"},
		"replicapool",
		"rook-ceph").Return(obj, nil)

	err := opt.Exec()
	assert.NoError(t, err)
}
