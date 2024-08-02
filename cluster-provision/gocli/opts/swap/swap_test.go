package swap

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestSwapOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SwapOpt Suite")
}

var _ = Describe("SwapOpt", func() {
	var (
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *swapOpt
	)

	BeforeEach(func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
		opt = NewSwapOpt(sshClient, 10, true, 1)
	})

	It("should execute SwapOpt successfully", func() {
		cmds := []string{
			"dd if=/dev/zero of=/swapfile count=" + fmt.Sprintf("%d", opt.size) + " bs=1G",
			"mkswap /swapfile",
			"swapon -a",
			"/bin/su -c \"echo vm.swappiness = " + fmt.Sprintf("%d", opt.swapiness) + " >> /etc/sysctl.conf\"",
			"sysctl vm.swappiness=" + fmt.Sprintf("%d", opt.swapiness),
			`sed -i 's/memorySwap: {}/memorySwap:\n  swapBehavior: UnlimitedSwap/g' /var/lib/kubelet/config.yaml`,
			"systemctl restart kubelet",
		}

		for _, cmd := range cmds {
			sshClient.EXPECT().Command(cmd)
		}

		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
