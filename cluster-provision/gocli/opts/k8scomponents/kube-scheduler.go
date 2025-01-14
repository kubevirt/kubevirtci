package k8scomponents

import (
	"fmt"
	"time"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

const componentKubeScheduler componentName = "kube-scheduler"

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

func (k *kubeSchedulerComponent) validateComponent() error {
	return validateComponentCommandHelper(k.sshClient, componentKubeScheduler)
}

func (k *kubeSchedulerComponent) configureComponent() error {
	if k.featureGates != "" {
		k.cmds = append(k.cmds, addFlagsToComponentCommand(componentKubeScheduler, fmt.Sprintf("--feature-gates=%s", k.featureGates)))
	}
	return runCommands(k.cmds, k.sshClient)
}

func (k *kubeSchedulerComponent) waitForComponentReady() error {
	return waitUntilReady(5*time.Minute, 3*time.Second, func() (bool, string) {
		return componentReadyHelper(componentKubeScheduler, k.sshClient, k.featureGates != "", false)
	})
}

func (k *kubeSchedulerComponent) requiredOnlyForMaster() bool {
	return true
}
