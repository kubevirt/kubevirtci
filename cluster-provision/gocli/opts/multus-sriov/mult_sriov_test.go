package multussriov

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

func TestMultusOpt(t *testing.T) {
	client := k8s.NewTestClient()
	opt := NewMultusSriovOpt(client)
	err := opt.Exec()
	assert.NoError(t, err)
}
