package nfscsi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

func TestNfsCsiOpt(t *testing.T) {
	updatePVC := func(action k8stesting.Action) (bool, runtime.Object, error) {
		createAction := action.(k8stesting.CreateAction)
		obj := createAction.GetObject().(*unstructured.Unstructured)
		status := map[string]interface{}{
			"phase": "Bound",
		}
		if err := unstructured.SetNestedField(obj.Object, status, "status"); err != nil {
			return true, nil, err
		}
		return false, obj, nil
	}

	r := k8s.NewReactorConfig("create", "persistentvolumeclaims", updatePVC)
	client := k8s.NewTestClient(r)
	opt := NewNfsCsiOpt(client)
	err := opt.Exec()
	assert.NoError(t, err)
}
