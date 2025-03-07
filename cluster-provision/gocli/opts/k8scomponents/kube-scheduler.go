package k8scomponents

import (
	"fmt"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type kubeSchedulerComponent struct {
	featureGates string
	cmds         []string
	sshClient    libssh.Client
}

func newKubeSchedulerComponent(sc libssh.Client, featureGates string) *kubeSchedulerComponent {
	return &kubeSchedulerComponent{
		sshClient:    sc,
		featureGates: featureGates,
	}
}

func (k *kubeSchedulerComponent) verifyComponent() error {
	return verifyComponentCommandHelper(k.sshClient, componentKubeScheduler)
}

func (k *kubeSchedulerComponent) prepareCommandsForConfiguration() error {
	if k.featureGates != "" {
		k.cmds = append(k.cmds, addFlagsToComponentCommand(componentKubeScheduler, fmt.Sprintf("--feature-gates=%s", k.featureGates)))
	}
	return nil
}

func (k *kubeSchedulerComponent) runCommandsToConfigure() error {
	return runCommands(k.cmds, k.sshClient)
}

func (k *kubeSchedulerComponent) waitForComponentReady() error {
	return componentReady(componentKubeScheduler, k.sshClient, k.featureGates != "", false)
}

func (k *kubeSchedulerComponent) requiredOnlyForMaster() bool {
	return true
}
