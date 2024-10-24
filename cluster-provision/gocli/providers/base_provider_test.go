package providers

import (
	"github.com/docker/docker/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/aaq"
	bindvfio "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/bind-vfio"
	etcdinmemory "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/etcd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/istio"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/labelnodes"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/nfscsi"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/node01"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/psa"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/rookceph"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

var _ = Describe("Node Provisioning", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		reactors  []k8s.ReactorConfig
		k8sClient k8s.K8sDynamicClient
		kp        *KubevirtProvider
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		reactors = []k8s.ReactorConfig{
			k8s.NewReactorConfig("create", "istiooperators", istio.IstioReactor),
			k8s.NewReactorConfig("create", "cephblockpools", rookceph.CephReactor),
			k8s.NewReactorConfig("create", "persistentvolumeclaims", nfscsi.NfsCsiReactor),
		}

		k8sClient = k8s.NewTestClient(reactors...)
		kp = NewKubevirtProvider("k8s-1.30", "", &client.Client{}, []KubevirtProviderOption{
			WithNodes(uint(1)),
			WithEnablePSA(true),
			WithEtcdCapacity("512M"),
			WithRunEtcdOnMemory(true),
			WithEnableCeph(true),
			WithEnablePrometheus(true),
			WithEnablePrometheusAlertManager(true),
			WithEnableIstio(true),
			WithAAQ(true),
			WithEnableNFSCSI(true),
			WithEnableGrafana(true),
		})
	})

	AfterEach(func() {
		mockCtrl.Finish()
		sshClient = nil
		k8sClient = nil
	})

	Describe("ProvisionNode", func() {
		It("should execute the correct commands", func() {
			etcdinmemory.AddExpectCalls(sshClient, "512M")
			bindvfio.AddExpectCalls(sshClient, "8086:2668")
			bindvfio.AddExpectCalls(sshClient, "8086:2415")
			psa.AddExpectCalls(sshClient)
			node01.AddExpectCalls(sshClient)

			err := kp.provisionNode(sshClient, 1)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ProvisionNodeK8sOpts", func() {
		It("should execute the correct K8s option commands", func() {
			kp.Client = k8sClient

			labelnodes.AddExpectCalls(sshClient, "node-role.kubernetes.io/control-plane")
			istio.AddExpectCalls(sshClient)
			aaq.AddExpectCalls(sshClient)

			err := kp.provisionK8sOpts(sshClient)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
