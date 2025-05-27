package k8scomponents

import (
	"fmt"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type kubeAPIComponent struct {
	runtimeConfig string
	featureGates  string
	cmds          []string
	sshClient     libssh.Client
}

func newKubeAPIComponent(sc libssh.Client, featureGates string, runtimeCofig string) *kubeAPIComponent {
	return &kubeAPIComponent{
		sshClient:     sc,
		featureGates:  featureGates,
		runtimeConfig: runtimeCofig,
	}
}

func (k *kubeAPIComponent) verifyComponent() error {
	return verifyComponentCommandHelper(k.sshClient, componentKubeAPIServer)
}

func (k *kubeAPIComponent) prepareCommandsForConfiguration() error {
	if k.featureGates != "" {
		k.cmds = append(k.cmds, addFlagsToComponentCommand(componentKubeAPIServer, fmt.Sprintf("--feature-gates=%s", k.featureGates)))
	}
	if k.runtimeConfig != "" {
		k.cmds = append(k.cmds, addFlagsToComponentCommand(componentKubeAPIServer, fmt.Sprintf("--runtime-config=%s", k.runtimeConfig)))
	}
	return nil
}

func (k *kubeAPIComponent) runCommandsToConfigure() error {
	return runCommands(k.cmds, k.sshClient)
}

func (k *kubeAPIComponent) waitForComponentReady() error {
	return componentReady(componentKubeAPIServer, k.sshClient, k.featureGates != "", k.runtimeConfig != "")
}

func (k *kubeAPIComponent) requiredOnlyForMaster() bool {
	return true
}
