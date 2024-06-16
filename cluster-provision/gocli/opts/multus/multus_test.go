package multus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestMultusOpt(t *testing.T) {
	mockK8sClient := kubevirtcimocks.NewMockK8sDynamicClient(gomock.NewController(t))
	multusOpt := NewMultusOpt(mockK8sClient)
	mockK8sClient.EXPECT().Apply(gomock.Any(), "manifests/multus.yaml").Return(nil)
	err := multusOpt.Exec()

	assert.NoError(t, err)
}
