package ksm

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestKsmOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KsmOpt Suite")
}

var _ = Describe("KsmOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *ksmOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewKsmOpt(sshClient, 20, 10)

		cmds := []string{
			"echo 1 | sudo tee /sys/kernel/mm/ksm/run >/dev/null",
			"echo " + fmt.Sprintf("%d", opt.scanInterval) + " | sudo tee /sys/kernel/mm/ksm/sleep_millisecs >/dev/null",
			"echo " + fmt.Sprintf("%d", opt.pagesToScan) + " | sudo tee /sys/kernel/mm/ksm/pages_to_scan >/dev/null",
		}

		for _, cmd := range cmds {
			sshClient.EXPECT().Command(cmd)
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute KsmOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
