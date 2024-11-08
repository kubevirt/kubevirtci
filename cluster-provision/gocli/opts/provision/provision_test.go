package provision

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestLinuxProvisioner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linux provision test suite")
}

var _ = Describe("Linux provision", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *linuxProvisioner
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewLinuxProvisioner(sshClient)
		AddExpectCalls(sshClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should provision linux successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
