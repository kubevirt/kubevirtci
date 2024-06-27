package cdi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

func TestCdiOpt(t *testing.T) {
	client := k8s.NewTestClient()
	opt := NewCdiOpt(client, "") // todo: cdi version
	err := opt.Exec()
	assert.NoError(t, err)
}
