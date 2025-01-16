package featuregate

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestFeatureGatesOpt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FeatureGatesOpt Suite")
}

type cmdOutput struct {
	output string
	err    error
}

var _ = Describe("featureGatesOpt", func() {
	var (
		sshClient *kubevirtcimocks.MockSSHClient
		opt       *featureGateOpt
	)

	BeforeEach(func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
		opt = NewFeatureGatesOpt(sshClient, []string{"FG1", "FG2"})
	})

	It("should execute FeatureGatesOpt successfully", func() {
		cmdsWithNoStdOut := map[string]cmdOutput{
			searchFeatureGatesInFile("/var/lib/kubelet/config.yaml"):                                               {"", nil},
			searchComponentsFilesCommand:                                                                           {"nonEmptyOutput", nil},
			searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeAPIServer)):     {"", nil},
			getComponentCommand(componentKubeAPIServer):                                                            {"feature-gate", nil},
			getComponentReadyContainers(componentKubeAPIServer):                                                    {"true", nil},
			searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeControllerMgr)): {"", nil},
			getComponentCommand(componentKubeControllerMgr):                                                        {"feature-gate", nil},
			getComponentReadyContainers(componentKubeControllerMgr):                                                {"true", nil},
			searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeScheduler)):     {"", nil},
			getComponentCommand(componentKubeScheduler):                                                            {"feature-gate", nil},
			getComponentReadyContainers(componentKubeScheduler):                                                    {"true", nil},
		}
		cmds := []string{
			addFeatureGatesFieldToKubeletConfigCommand,
			addFeatureGatesToKubeletConfigCommand("FG1"),
			addFeatureGatesToKubeletConfigCommand("FG2"),
			kubeletRestartCommand,
			addFeatureGatesToComponentCommand(componentKubeAPIServer, []string{"FG1", "FG2"}),
			addFeatureGatesToComponentCommand(componentKubeControllerMgr, []string{"FG1", "FG2"}),
			addFeatureGatesToComponentCommand(componentKubeScheduler, []string{"FG1", "FG2"}),
		}

		for cmd, output := range cmdsWithNoStdOut {
			sshClient.EXPECT().CommandWithNoStdOut(cmd).Return(output.output, output.err)
		}

		for _, cmd := range cmds {
			sshClient.EXPECT().Command(cmd)
		}

		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})

	It("should fail when feature gate already exists in kubelet", func() {
		sshClient.EXPECT().CommandWithNoStdOut(searchFeatureGatesInFile("/var/lib/kubelet/config.yaml")).Return("featureGates", nil)

		err := opt.Exec()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(featureGateExistInKubeletError))
	})

	It("should fail when feature gate already exists in components", func() {
		sshClient.EXPECT().CommandWithNoStdOut(searchFeatureGatesInFile("/var/lib/kubelet/config.yaml")).Return("", nil)
		sshClient.EXPECT().Command(addFeatureGatesFieldToKubeletConfigCommand).Return(nil)
		sshClient.EXPECT().Command(addFeatureGatesToKubeletConfigCommand("FG1")).Return(nil)
		sshClient.EXPECT().Command(addFeatureGatesToKubeletConfigCommand("FG2")).Return(nil)
		sshClient.EXPECT().Command(kubeletRestartCommand).Return(nil)
		sshClient.EXPECT().CommandWithNoStdOut(searchComponentsFilesCommand).Return("nonEmptyOutput", nil)
		sshClient.EXPECT().CommandWithNoStdOut(searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeAPIServer))).Return("featureGates", nil)

		err := opt.Exec()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(featureGateExistInComponentCommandErrorFormat, componentKubeAPIServer)))
	})

	It("should retry when API server does not respond or if changes are not propagated to components", func() {
		cmdsWithNoStdOut := map[string]cmdOutput{
			searchFeatureGatesInFile("/var/lib/kubelet/config.yaml"):                                               {"", nil},
			searchComponentsFilesCommand:                                                                           {"nonEmptyOutput", nil},
			searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeAPIServer)):     {"", nil},
			searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeControllerMgr)): {"", nil},
			searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeScheduler)):     {"", nil},
		}
		cmds := []string{
			addFeatureGatesFieldToKubeletConfigCommand,
			addFeatureGatesToKubeletConfigCommand("FG1"),
			addFeatureGatesToKubeletConfigCommand("FG2"),
			kubeletRestartCommand,
			addFeatureGatesToComponentCommand(componentKubeAPIServer, []string{"FG1", "FG2"}),
			addFeatureGatesToComponentCommand(componentKubeControllerMgr, []string{"FG1", "FG2"}),
			addFeatureGatesToComponentCommand(componentKubeScheduler, []string{"FG1", "FG2"}),
		}

		for cmd, output := range cmdsWithNoStdOut {
			sshClient.EXPECT().CommandWithNoStdOut(cmd).Return(output.output, output.err)
		}

		for _, cmd := range cmds {
			sshClient.EXPECT().Command(cmd)
		}

		// First 2 attempts fail
		sshClient.EXPECT().CommandWithNoStdOut(getComponentCommand(componentKubeAPIServer)).Return("some-output", nil).Times(1)
		sshClient.EXPECT().CommandWithNoStdOut(getComponentCommand(componentKubeAPIServer)).Return("", fmt.Errorf("API server not responding")).Times(1)
		// Third attempt succeeds
		sshClient.EXPECT().CommandWithNoStdOut(getComponentCommand(componentKubeAPIServer)).Return("feature-gate", nil).Times(1)
		sshClient.EXPECT().CommandWithNoStdOut(getComponentReadyContainers(componentKubeAPIServer)).Return("true", nil)

		sshClient.EXPECT().CommandWithNoStdOut(getComponentCommand(componentKubeControllerMgr)).Return("feature-gate", nil)
		sshClient.EXPECT().CommandWithNoStdOut(getComponentReadyContainers(componentKubeControllerMgr)).Return("true", nil)
		sshClient.EXPECT().CommandWithNoStdOut(getComponentCommand(componentKubeScheduler)).Return("feature-gate", nil)
		sshClient.EXPECT().CommandWithNoStdOut(getComponentReadyContainers(componentKubeScheduler)).Return("true", nil)

		err := opt.Exec()
		Expect(err).NotTo(HaveOccurred())
	})
})
