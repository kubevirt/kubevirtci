package node01

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/common"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/extranics"
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
		opt = NewNode01Provisioner(sshClient, false, false, false, extranics.Config{})
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute Node01Provisioner successfully", func() {
		AddExpectCalls(sshClient)
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should stop before kubeadm init when hostname readiness fails", func() {
		gomock.InOrder(
			sshClient.EXPECT().Command(gomock.Any()),
			sshClient.EXPECT().Command(common.WaitForHostnameCommand("node01")).Return(errors.New("hostname not ready")),
		)
		Expect(opt.Exec()).To(MatchError(ContainSubstring("hostname not ready")))
	})
})
