package etcdinmemory

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestEtcdOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EtcdOpt Suite")
}

var _ = Describe("EtcdOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *etcdInMemOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewEtcdInMemOpt(sshClient, "512M")
		AddExpectCalls(sshClient, opt.etcdSize)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute EtcdInMemOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
