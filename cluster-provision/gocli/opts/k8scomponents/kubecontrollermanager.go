package k8scomponents

import (
	"fmt"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type kubeControllerMgrComponent struct {
	featureGates string
	cmds         []string
	sshClient    libssh.Client
}

func newKubeControllerMgrComponent(sc libssh.Client, featureGates string) *kubeControllerMgrComponent {
	return &kubeControllerMgrComponent{
		sshClient:    sc,
		featureGates: featureGates,
	}
}

func (k *kubeControllerMgrComponent) verifyComponent() error {
	return verifyComponentCommandHelper(k.sshClient, componentKubeControllerMgr)
}

func (k *kubeControllerMgrComponent) prepareCommandsForConfiguration() error {
	if k.featureGates != "" {
		k.cmds = append(k.cmds, addFlagsToComponentCommand(componentKubeControllerMgr, fmt.Sprintf("--feature-gates=%s", k.featureGates)))
	}
	return nil
}

func (k *kubeControllerMgrComponent) runCommandsToConfigure() error {
	return runCommands(k.cmds, k.sshClient)
}

func (k *kubeControllerMgrComponent) waitForComponentReady() error {
	return componentReady(componentKubeControllerMgr, k.sshClient, k.featureGates != "", false)
}

func (k *kubeControllerMgrComponent) requiredOnlyForMaster() bool {
	return true
}
