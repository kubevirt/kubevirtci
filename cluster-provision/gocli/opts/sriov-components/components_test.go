package sriovcomponents

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

func TestSriovComponents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SriovComponents Suite")
}

var _ = Describe("SriovComponents", func() {
	var (
		mockCtrl  *gomock.Controller
		k8sClient k8s.K8sDynamicClient
		opt       *sriovComponentsOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		k8sClient = k8s.NewTestClient()
		opt = NewSriovComponentsOpt(k8sClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute sriov components successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
