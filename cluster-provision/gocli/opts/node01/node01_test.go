package node01

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestNodeProvisionerOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node01Provisioner Suite")
}

var _ = Describe("Node01Provisioner", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *node01Provisioner
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewNode01Provisioner(sshClient, false, false, false, false)
		AddExpectCalls(sshClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute Node01Provisioner successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
