package cdi

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestCdiOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CdiOpt Suite")
}

var _ = Describe("CdiOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		client    k8s.K8sDynamicClient
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *cdiOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		client = k8s.NewTestClient()
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewCdiOpt(client, sshClient, "")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute CdiOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
