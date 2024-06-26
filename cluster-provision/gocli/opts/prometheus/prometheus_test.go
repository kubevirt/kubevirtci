package prometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

func TestPrometheusOpt(t *testing.T) {
	client := k8s.NewTestClient()
	opt := NewPrometheusOpt(client, true, true)
	err := opt.Exec()
	assert.NoError(t, err)
}
