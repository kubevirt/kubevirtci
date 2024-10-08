package labelnodes

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestNodeLabel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NodeLabeler test suite")
}

var _ = Describe("NodeLabeler", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *nodeLabler
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewNodeLabler(sshClient, "node-role.kubernetes.io/control-plane")
		AddExpectCalls(sshClient, "node-role.kubernetes.io/control-plane")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should label nodes successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
