package k8scomponents

import (
	"fmt"
	"time"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

const componentKubeAPIServer componentName = "kube-apiserver"

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

func (k *kubeAPIComponent) validateComponent() error {
	return validateComponentCommandHelper(k.sshClient, componentKubeAPIServer)
}

func (k *kubeAPIComponent) configureComponent() error {
	if k.featureGates != "" {
		k.cmds = append(k.cmds, addFlagsToComponentCommand(componentKubeAPIServer, fmt.Sprintf("--feature-gates=%s", k.featureGates)))
	}
	if k.runtimeConfig != "" {
		k.cmds = append(k.cmds, addFlagsToComponentCommand(componentKubeAPIServer, fmt.Sprintf("--runtime-config=%s", k.runtimeConfig)))
	}
	return runCommands(k.cmds, k.sshClient)
}

func (k *kubeAPIComponent) waitForComponentReady() error {
	return waitUntilReady(5*time.Minute, 3*time.Second, func() (bool, string) {
		return componentReadyHelper(componentKubeAPIServer, k.sshClient, k.featureGates != "", k.runtimeConfig != "")
	})
}

func (k *kubeAPIComponent) requiredOnlyForMaster() bool {
	return true
}
