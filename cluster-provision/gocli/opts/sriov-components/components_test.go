package sriovcomponents

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

func TestSriovComponents(t *testing.T) {
	client := k8s.NewTestClient()
	opt := NewSriovComponentsOpt(client)
	err := opt.Exec()
	assert.NoError(t, err)
}
