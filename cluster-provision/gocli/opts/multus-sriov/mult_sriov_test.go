package multussriov

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

func TestMultusSriovOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MultusSriovOpt Suite")
}

var _ = Describe("MultusSriovOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		k8sClient k8s.K8sDynamicClient
		opt       *multusSriovOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		k8sClient = k8s.NewTestClient()
		opt = NewMultusSriovOpt(k8sClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute MultusSriovOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
