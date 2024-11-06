package nodes

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestNodeProvisionerOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NodesProvisioner Suite")
}

var _ = Describe("NodesProvisioner", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *nodesProvisioner
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewNodesProvisioner(sshClient, false, false)
		AddExpectCalls(sshClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute NodesProvisioner successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
