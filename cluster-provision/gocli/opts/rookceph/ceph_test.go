package rookceph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

func TestWithFakeClient(t *testing.T) {
	r := k8s.NewReactorConfig("create", "cephblockpools", CephReactor)
	testClient := k8s.NewTestClient(r)
	opt := NewCephOpt(testClient)
	err := opt.Exec()
	assert.NoError(t, err)
}
