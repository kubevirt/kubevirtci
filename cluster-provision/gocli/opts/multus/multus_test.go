package multus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

func TestMultusOpt(t *testing.T) {
	client := k8s.NewTestClient()
	opt := NewMultusOpt(client)
	err := opt.Exec()
	assert.NoError(t, err)
}
