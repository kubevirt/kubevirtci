package cnao

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestCnaoOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CnaoOpt Suite")
}

var _ = Describe("CnaoOpt", func() {
	var (
		mockCtrl      *gomock.Controller
		client        k8s.K8sDynamicClient
		sshClient     *kubevirtcimocks.MockSSHClient
		opt           *cnaoOpt
		skipCR        bool
		dncEnabled    bool
		multusEnabled bool
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		client = k8s.NewTestClient()
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should execute create CNAO with Multus", func() {
		skipCR = false
		dncEnabled = true
		multusEnabled = false

		opt = NewCnaoOpt(client, sshClient, multusEnabled, dncEnabled, skipCR)

		sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s")
		opt.Exec()

		obj, err := client.Get(schema.GroupVersionKind{Group: "networkaddonsoperator.network.kubevirt.io",
			Version: "v1",
			Kind:    "NetworkAddonsConfig"}, "cluster", "")
		Expect(err).NotTo(HaveOccurred())

		spec, ok := obj.Object["spec"].(map[string]interface{})
		Expect(ok).To(Equal(true))
		Expect(spec).To(HaveKey("multus"))
		Expect(spec).To(HaveKey("multusDynamicNetworks"))
	})

	It("should execute create CNAO without Multus", func() {
		skipCR = false
		dncEnabled = false
		multusEnabled = true

		opt = NewCnaoOpt(client, sshClient, multusEnabled, dncEnabled, skipCR)
		sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s")
		opt.Exec()

		obj, err := client.Get(schema.GroupVersionKind{Group: "networkaddonsoperator.network.kubevirt.io",
			Version: "v1",
			Kind:    "NetworkAddonsConfig"}, "cluster", "")
		Expect(err).NotTo(HaveOccurred())

		spec, ok := obj.Object["spec"].(map[string]interface{})
		Expect(ok).To(Equal(true))
		Expect(spec).NotTo(HaveKey("multus"))
		Expect(spec).NotTo(HaveKey("multusDynamicNetworks"))
	})

	It("should execute create CNAO with dynamic networks controller", func() {
		skipCR = false
		dncEnabled = true
		multusEnabled = false

		opt = NewCnaoOpt(client, sshClient, multusEnabled, dncEnabled, skipCR)
		sshClient.EXPECT().Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s")
		opt.Exec()

		obj, err := client.Get(schema.GroupVersionKind{Group: "networkaddonsoperator.network.kubevirt.io",
			Version: "v1",
			Kind:    "NetworkAddonsConfig"}, "cluster", "")
		Expect(err).NotTo(HaveOccurred())

		spec, ok := obj.Object["spec"].(map[string]interface{})
		Expect(ok).To(Equal(true))
		Expect(spec).To(HaveKey("multusDynamicNetworks"))
		Expect(spec).To(HaveKey("multus"))
	})
})
