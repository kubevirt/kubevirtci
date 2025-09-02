package k8scomponents

import (
	"fmt"
	"time"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

const componentKubeControllerMgr componentName = "kube-controller-manager"

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

func (k *kubeControllerMgrComponent) validateComponent() error {
	return validateComponentCommandHelper(k.sshClient, componentKubeControllerMgr)
}

func (k *kubeControllerMgrComponent) configureComponent() error {
	if k.featureGates != "" {
		k.cmds = append(k.cmds, addFlagsToComponentCommand(componentKubeControllerMgr, fmt.Sprintf("--feature-gates=%s", k.featureGates)))
	}
	return runCommands(k.cmds, k.sshClient)
}

func (k *kubeControllerMgrComponent) waitForComponentReady() error {
	return waitUntilReady(5*time.Minute, 3*time.Second, func() (bool, string) {
		return componentReadyHelper(componentKubeControllerMgr, k.sshClient, k.featureGates != "", false)
	})
}

func (k *kubeControllerMgrComponent) requiredOnlyForMaster() bool {
	return true
}
