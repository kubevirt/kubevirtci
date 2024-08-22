package cmd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/nodesconfig"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/aaq"
	bindvfio "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/bind-vfio"
	etcdinmemory "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/etcd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/istio"
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
	})

	AfterEach(func() {
		mockCtrl.Finish()
		sshClient = nil
		k8sClient = nil
	})

	Describe("ProvisionNode", func() {
		It("should execute the correct commands", func() {
			linuxConfigFuncs := []nodesconfig.LinuxConfigFunc{
				nodesconfig.WithEtcdInMemory(true),
				nodesconfig.WithEtcdSize("512M"),
				nodesconfig.WithPSA(true),
			}

			n := nodesconfig.NewNodeLinuxConfig(1, "k8s-1.30", linuxConfigFuncs)

			etcdinmemory.AddExpectCalls(sshClient, "512M")
			bindvfio.AddExpectCalls(sshClient, "8086:2668")
			bindvfio.AddExpectCalls(sshClient, "8086:2415")
			psa.AddExpectCalls(sshClient)
			node01.AddExpectCalls(sshClient)

			err := provisionNode(sshClient, n)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ProvisionNodeK8sOpts", func() {
		It("should execute the correct K8s option commands", func() {
			k8sConfs := []nodesconfig.K8sConfigFunc{
				nodesconfig.WithCeph(true),
				nodesconfig.WithPrometheus(true),
				nodesconfig.WithAlertmanager(true),
				nodesconfig.WithGrafana(true),
				nodesconfig.WithIstio(true),
				nodesconfig.WithNfsCsi(true),
				nodesconfig.WithAAQ(true),
			}
			n := nodesconfig.NewNodeK8sConfig(k8sConfs)

			istio.AddExpectCalls(sshClient)
			aaq.AddExpectCalls(sshClient)

			err := provisionK8sOptions(sshClient, k8sClient, n, "k8s-1.30")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
