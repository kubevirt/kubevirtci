package nri

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestEtcdOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Network Resources Injector Suite")
}

var _ = Describe("Network Resources Injector", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		k8sClient k8s.K8sDynamicClient
		opt       *networkResourcesInjectorOpt
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		k8sClient = k8s.NewTestClient()
		opt = NewNetworkResourcesInjectorOpt(sshClient, k8sClient)
		sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf rollout status -n kube-system deploy/network-resources-injector --timeout=200s")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute NRI successfully", func() {
		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
