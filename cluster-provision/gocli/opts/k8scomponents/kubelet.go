package k8scomponents

import (
	"fmt"
	"strings"
	"time"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
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

func (k *kubeletComponent) verifyComponent() error {
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

func (k *kubeletComponent) prepareCommandsForConfiguration() error {
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

	return nil
}

func (k *kubeletComponent) runCommandsToConfigure() error {
	return runCommands(k.cmds, k.sshClient)
}

func (k *kubeletComponent) waitForComponentReady() error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	timeoutChannel := time.After(3 * time.Minute)
	select {
	case <-timeoutChannel:
		return fmt.Errorf("timed out after 3 minutes waiting for node to be ready")
	case <-ticker.C:
		output, err := k.sshClient.CommandWithNoStdOut(getNodeReadyStatusCommand)
		if err == nil && !strings.Contains(output, "false") {
			return nil
		}
		if err != nil {
			fmt.Printf("Modifying kubelet configuration, API server not responding yet, err: %v\n", err)
		}
	}
	return nil
}

func (k *kubeletComponent) requiredOnlyForMaster() bool {
	return false
}
