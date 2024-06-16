package nfscsi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

func TestNfsCsiOpt(t *testing.T) {
	client := k8s.NewTestClient()
	opt := NewNfsCsiOpt(client)
	err := opt.Exec()
	assert.NoError(t, err)
}
