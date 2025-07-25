package k8scomponents

import (
	"fmt"
	"strings"
	"time"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

const (
	featureGateExistInKubeletError             = "feature gates should not exist in kubelet by default"
	kubeletRestartCommand                      = "sudo systemctl restart kubelet"
	addFeatureGatesFieldToKubeletConfigCommand = "sudo echo -e 'featureGates:' >> /var/lib/kubelet/config.yaml"
)

type kubeletComponent struct {
	featureGates string
	cmds         []string
	sshClient    libssh.Client
}

func newKubeletComponent(sc libssh.Client, featureGates string) *kubeletComponent {
	return &kubeletComponent{
		sshClient:    sc,
		featureGates: featureGates,
	}
}

func (k *kubeletComponent) validateComponent() error {
	output, err := k.sshClient.CommandWithNoStdOut(searchFeatureGatesInFile("/var/lib/kubelet/config.yaml"))
	if err != nil {
		return err
	}
	fgsExist := len(output) > 0
	if fgsExist {
		return fmt.Errorf(featureGateExistInKubeletError)
	}
	return nil
}

func (k *kubeletComponent) configureComponent() error {
	if k.featureGates != "" {
		k.cmds = append(k.cmds, addFeatureGatesFieldToKubeletConfigCommand)
		keyValuePairs := strings.Split(k.featureGates, ",")
		var formattedFeatureGates []string
		for _, pair := range keyValuePairs {
			formattedFeatureGates = append(formattedFeatureGates, strings.Replace(pair, "=", ": ", 1))
		}
		for _, fg := range formattedFeatureGates {
			k.cmds = append(k.cmds, addFeatureGatesToKubeletConfigCommand(fg))
		}
		k.cmds = append(k.cmds, kubeletRestartCommand)
	}
	return runCommands(k.cmds, k.sshClient)
}

func (k *kubeletComponent) waitForComponentReady() error {
	return waitUntilReady(3*time.Minute, 1*time.Second, func() (bool, string) {
		output, err := k.sshClient.CommandWithNoStdOut(getNodeReadyStatusCommand)
		if err == nil && !strings.Contains(output, "false") {
			return true, ""
		}
		if err != nil {
			return false, fmt.Sprintf("Modifying kubelet configuration, API server not responding yet, err: %v", err)
		}
		return false, ""
	})
}

func (k *kubeletComponent) requiredOnlyForMaster() bool {
	return false
}

func addFeatureGatesToKubeletConfigCommand(feature string) string {
	addFeatureGatesToKubeletConfigCommandFormatFormat := `sudo sed -i 's/featureGates:/featureGates:\n  %s/g' /var/lib/kubelet/config.yaml`
	return fmt.Sprintf(addFeatureGatesToKubeletConfigCommandFormatFormat, feature)
}
