package k8scomponents

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
	var sshClient *kubevirtcimocks.MockSSHClient

	BeforeEach(func() {
		sshClient = kubevirtcimocks.NewMockSSHClient(gomock.NewController(GinkgoT()))
	})

	DescribeTable("when ", func(featureGates string, runtimeConfig string, cmds []string, cmdsWithOutPut map[string]cmdOutput, expectedMsg string) {
		opt := NewK8sComponentsOpt(sshClient, featureGates, runtimeConfig)
		for cmd, output := range cmdsWithOutPut {
			sshClient.EXPECT().CommandWithNoStdOut(cmd).Return(output.output, output.err)
		}

		for _, cmd := range cmds {
			sshClient.EXPECT().Command(cmd)
		}

		err := opt.Exec()
		if expectedMsg != "" {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedMsg))
		} else {
			Expect(err).NotTo(HaveOccurred())
		}
	},
		Entry("should execute FeatureGatesOpt successfully with only featureGates",
			"FG1=true,FG2=true", "",
			[]string{addFeatureGatesFieldToKubeletConfigCommand,
				addFeatureGatesToKubeletConfigCommand("FG1: true"),
				addFeatureGatesToKubeletConfigCommand("FG2: true"),
				kubeletRestartCommand,
				addFlagsToComponentCommand(componentKubeAPIServer, "--feature-gates=FG1=true,FG2=true"),
				addFlagsToComponentCommand(componentKubeControllerMgr, "--feature-gates=FG1=true,FG2=true"),
				addFlagsToComponentCommand(componentKubeScheduler, "--feature-gates=FG1=true,FG2=true"),
			}, map[string]cmdOutput{
				searchFeatureGatesInFile("/var/lib/kubelet/config.yaml"): {"", nil},
				getNodeReadyStatusCommand:                                {"true", nil},
				searchComponentsFilesCommand:                             {"nonEmptyOutput", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeAPIServer)):     {"", nil},
				getComponentCommand(componentKubeAPIServer):                                                            {"feature-gate", nil},
				getComponentReadyContainers(componentKubeAPIServer):                                                    {"true", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeControllerMgr)): {"", nil},
				getComponentCommand(componentKubeControllerMgr):                                                        {"feature-gate", nil},
				getComponentReadyContainers(componentKubeControllerMgr):                                                {"true", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeScheduler)):     {"", nil},
				getComponentCommand(componentKubeScheduler):                                                            {"feature-gate", nil},
				getComponentReadyContainers(componentKubeScheduler):                                                    {"true", nil},
			}, ""),
		Entry("should execute FeatureGatesOpt successfully with only runtimeConfig",
			"", "runtimeConfig",
			[]string{addFlagsToComponentCommand(componentKubeAPIServer, "--runtime-config=runtimeConfig")},
			map[string]cmdOutput{
				searchFeatureGatesInFile("/var/lib/kubelet/config.yaml"): {"", nil},
				getNodeReadyStatusCommand:                                {"true", nil},
				searchComponentsFilesCommand:                             {"nonEmptyOutput", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeAPIServer)):     {"", nil},
				getComponentCommand(componentKubeAPIServer):                                                            {"runtime-config", nil},
				getComponentReadyContainers(componentKubeAPIServer):                                                    {"true", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeControllerMgr)): {"", nil},
				getComponentReadyContainers(componentKubeControllerMgr):                                                {"true", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeScheduler)):     {"", nil},
				getComponentReadyContainers(componentKubeScheduler):                                                    {"true", nil},
			}, ""),
		Entry("should fail when kubelet feature gate already exists",
			"FG1=true,FG2=true", "",
			[]string{}, map[string]cmdOutput{
				searchComponentsFilesCommand:                             {"nonEmptyOutput", nil},
				searchFeatureGatesInFile("/var/lib/kubelet/config.yaml"): {"featureGates", nil},
			}, featureGateExistInKubeletError),
		Entry("should fail when feature gate already exists in components",
			"FG1=true,FG2=true", "runtimeConfig",
			[]string{
				addFeatureGatesFieldToKubeletConfigCommand,
				addFeatureGatesToKubeletConfigCommand("FG1: true"),
				addFeatureGatesToKubeletConfigCommand("FG2: true"),
				kubeletRestartCommand,
			},
			map[string]cmdOutput{
				searchComponentsFilesCommand:                             {"nonEmptyOutput", nil},
				getNodeReadyStatusCommand:                                {"true", nil},
				searchFeatureGatesInFile("/var/lib/kubelet/config.yaml"): {"", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeAPIServer)): {"featureGates", nil},
			}, fmt.Sprintf(featureGateExistInComponentCommandErrorFormat, componentKubeAPIServer)),
		Entry("should retry when API server does not respond or if changes are not propagated to components",
			"FG1=true,FG2=true", "",
			[]string{
				addFeatureGatesFieldToKubeletConfigCommand,
				addFeatureGatesToKubeletConfigCommand("FG1: true"),
				addFeatureGatesToKubeletConfigCommand("FG2: true"),
				kubeletRestartCommand,
				addFlagsToComponentCommand(componentKubeAPIServer, "--feature-gates=FG1=true,FG2=true"),
				addFlagsToComponentCommand(componentKubeControllerMgr, "--feature-gates=FG1=true,FG2=true"),
				addFlagsToComponentCommand(componentKubeScheduler, "--feature-gates=FG1=true,FG2=true"),
			}, map[string]cmdOutput{
				searchComponentsFilesCommand:                             {"nonEmptyOutput", nil},
				searchFeatureGatesInFile("/var/lib/kubelet/config.yaml"): {"", nil},
				getNodeReadyStatusCommand:                                {"true", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeAPIServer)):     {"", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeControllerMgr)): {"", nil},
				searchFeatureGatesInFile(fmt.Sprintf("/etc/kubernetes/manifests/%s.yaml", componentKubeScheduler)):     {"", nil},
				// First 2 attempts fail
				getComponentCommand(componentKubeAPIServer): {"some-output", nil},
				getComponentCommand(componentKubeAPIServer): {"", fmt.Errorf("API server not responding")},
				// Third attempt succeeds
				getComponentCommand(componentKubeAPIServer):             {"feature-gate", nil},
				getComponentReadyContainers(componentKubeAPIServer):     {"true", nil},
				getComponentCommand(componentKubeControllerMgr):         {"feature-gate", nil},
				getComponentReadyContainers(componentKubeControllerMgr): {"true", nil},
				getComponentCommand(componentKubeScheduler):             {"feature-gate", nil},
				getComponentReadyContainers(componentKubeScheduler):     {"true", nil},
			}, ""))
})
