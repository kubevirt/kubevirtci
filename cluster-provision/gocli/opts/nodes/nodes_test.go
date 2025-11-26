package nodes

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestNodeProvisionerOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NodesProvisioner Suite")
}

var _ = Describe("nodes", func() {
	When("NodesProvisioner", func() {
		var (
			mockCtrl  *gomock.Controller
			sshClient *kubevirtcimocks.MockSSHClient
			opt       *nodesProvisioner
		)

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
			opt = NewNodesProvisioner("k8s-1.32", sshClient, false)
			AddExpectCalls(sshClient)
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		It("should execute NodesProvisioner successfully", func() {
			err := opt.Exec()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	DescribeTable("calling featureGateFlag",
		func(k8sVersion, expectedValue string) {
			np := NewNodesProvisioner(k8sVersion, nil, false)
			Expect(np.featureGatesFlag()).To(BeEquivalentTo(expectedValue))
		},
		Entry("should not add new fg if 1.32", "k8s-1.32", "--feature-gates=NodeSwap=true"),
	)

	When("job name does not contain version", func() {
		DescribeTable("calling featureGateFlag",
			func(k8sVersion, expectedValue string) {
				kvProviderOrig, kvProviderDefined := os.LookupEnv(kubevirtProviderEnv)

				Expect(os.Setenv(kubevirtProviderEnv, k8sVersion)).To(Succeed())
				DeferCleanup(func() {
					if kvProviderDefined {
						Expect(os.Setenv(kubevirtProviderEnv, kvProviderOrig)).To(Succeed())
					} else {
						Expect(os.Unsetenv(kubevirtProviderEnv)).To(Succeed())
					}
				})

				np := NewNodesProvisioner("name-with-no-version", nil, false)
				Expect(np.featureGatesFlag()).To(BeEquivalentTo(expectedValue))
			},
			Entry("should not add new fg if 1.32", "k8s-1.32", "--feature-gates=NodeSwap=true"),
		)
	})
})
