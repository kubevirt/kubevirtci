package psa

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRealTimeOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PsaOpt Suite")
}

var _ = Describe("PsaOpt", func() {
	var (
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *psaOpt
	)

	BeforeEach(func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
		opt = NewPsaOpt(sshClient)
		AddExpectCalls(sshClient)
	})

	It("should execute PsaOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
