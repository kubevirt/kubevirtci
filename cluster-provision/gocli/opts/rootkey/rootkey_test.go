package rootkey

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRootKey(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RootKey Suite")
}

var _ = Describe("RootKey", func() {
	var (
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *rootKey
	)

	BeforeEach(func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
		opt = NewRootKey(sshClient)
	})

	It("should execute RootKey successfully", func() {
		sshClient.EXPECT().Command("echo '" + string(key) + "' | sudo tee /root/.ssh/authorized_keys > /dev/null")
		sshClient.EXPECT().Command("sudo service sshd restart")

		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
