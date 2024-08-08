package rookceph

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/testing"
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
