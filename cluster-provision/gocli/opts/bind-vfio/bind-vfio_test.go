package bindvfio

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestBindVfio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BindVfio Suite")
}

var _ = Describe("BindVfio", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *bindVfioOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewBindVfioOpt(sshClient, "8086:2668")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute BindVfio successfully", func() {
		AddExpectCalls(sshClient, opt.pciID)
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
