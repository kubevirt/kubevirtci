package realtime

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestRealTimeOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RealtimeOpt Suite")
}

var _ = Describe("RealtimeOpt", func() {
	var (
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *realtimeOpt
	)

	BeforeEach(func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
		opt = NewRealtimeOpt(sshClient)
	})

	It("should execute RealtimeOpt successfully", func() {
		sshClient.EXPECT().Command("echo kernel.sched_rt_runtime_us=-1 > /etc/sysctl.d/realtime.conf")
		sshClient.EXPECT().Command("sysctl --system")

		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
