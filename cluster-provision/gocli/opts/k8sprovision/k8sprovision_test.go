package k8sprovision

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestK8sProvision(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "K8s provision test suite")
}

var _ = Describe("K8s provision", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *k8sProvisioner
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt := NewK8sProvisioner(sshClient, "1.30", true)
		AddExpectCalls(sshClient, opt.version, opt.slim)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should provision k8s successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
