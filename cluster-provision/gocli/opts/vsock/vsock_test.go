package vsock

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestVsockOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VsockOpt Suite")
}

var _ = Describe("VsockOpt", func() {
	var sshClient *kubevirtcimocks.MockSSHClient

	It("should execute VsockOpt successfully with global mode", func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
		opt, err := NewVsockOpt(sshClient, "global")
		Expect(err).NotTo(HaveOccurred())

		cmds := []string{
			"modprobe vsock",
			"sysctl --write net.vsock.child_ns_mode=global",
		}

		for _, cmd := range cmds {
			sshClient.EXPECT().Command(cmd)
		}

		Expect(opt.Exec()).To(Succeed())
	})

	It("should execute VsockOpt successfully with local mode", func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
		opt, err := NewVsockOpt(sshClient, "local")
		Expect(err).NotTo(HaveOccurred())

		cmds := []string{
			"modprobe vsock",
			"sysctl --write net.vsock.child_ns_mode=local",
		}

		for _, cmd := range cmds {
			sshClient.EXPECT().Command(cmd)
		}

		Expect(opt.Exec()).To(Succeed())
	})

	It("should fail with invalid mode", func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
		_, err := NewVsockOpt(sshClient, "invalid")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid vsock child namespace mode"))
	})
})
