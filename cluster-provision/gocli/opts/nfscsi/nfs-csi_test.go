package nfscsi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

func TestNfsCsiOpt(t *testing.T) {
	r := k8s.NewReactorConfig("create", "persistentvolumeclaims", NfsCsiReactor)
	client := k8s.NewTestClient(r)
	opt := NewNfsCsiOpt(client)
	err := opt.Exec()
	assert.NoError(t, err)
}
