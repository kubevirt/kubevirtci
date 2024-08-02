package cnao

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestCnaoOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CnaoOpt Suite")
}

var _ = Describe("CnaoOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		client    k8s.K8sDynamicClient
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *cnaoOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		client = k8s.NewTestClient()
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		opt = NewCnaoOpt(client, sshClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute CnaoOpt successfully", func() {
		sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s")
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
