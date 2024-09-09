package rookceph

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestNfsCsiOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CephOpt Suite")
}

var _ = Describe("CephOpt", func() {
	var (
		k8sClient k8s.K8sDynamicClient
		opt       *cephOpt
		sshClient *kubevirtcimocks.MockSSHClient
	)

	BeforeEach(func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
		AddExpectCalls(sshClient)
		r := k8s.NewReactorConfig("create", "cephblockpools", CephReactor)
		k8sClient = k8s.NewTestClient(r)
		opt = NewCephOpt(k8sClient, sshClient)
	})

	It("should execute Ceph successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
