package multus

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestMultusOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MultusOpt Suite")
}

var _ = Describe("MultusOpt", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		k8sClient k8s.K8sDynamicClient
		opt       *multusOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		k8sClient = k8s.NewTestClient()
		opt = NewMultusOpt(k8sClient, sshClient)

		sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf rollout status -n kube-system ds/kube-multus-ds --timeout=200s")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute MultusOpt successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
