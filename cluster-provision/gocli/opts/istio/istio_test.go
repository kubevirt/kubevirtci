package istio

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestIstioOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IstioOpt Suite")
}

var _ = Describe("IstioOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		k8sclient k8s.K8sDynamicClient
		opt       *istioOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		k8sclient = k8s.NewTestClient()
		opt = NewIstioOpt(sshClient, k8sclient, false)
		AddExpectCalls(sshClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute IstioOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
