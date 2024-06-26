package cnao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

func TestCnaoOpt(t *testing.T) {
	client := k8s.NewTestClient()
	opt := NewCnaoOpt(client)
	err := opt.Exec()
	assert.NoError(t, err)
}
