package k8scomponents

import (
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type k8sComponentsOpt struct {
	sshClient     libssh.Client
	kubeletSetup  bool
	runtimeConfig string
	featureGates  string
	component     []component
}

func NewK8sComponentsOpt(sc libssh.Client, featureGates string, runtimeConfig string) *k8sComponentsOpt {
	return &k8sComponentsOpt{
		sshClient:     sc,
		runtimeConfig: runtimeConfig,
		featureGates:  featureGates,
		component: []component{
			newKubeletComponent(sc, featureGates),
			newKubeAPIComponent(sc, featureGates, runtimeConfig),
			newKubeSchedulerComponent(sc, featureGates),
			newKubeControllerMgrComponent(sc, featureGates),
		},
	}
}

func (o *k8sComponentsOpt) Exec() error {
	output, err := o.sshClient.CommandWithNoStdOut(searchComponentsFilesCommand)
	if err != nil {
		return err
	}
	onMasterNode := len(output) > 0
	for _, c := range o.component {
		if c.requiredOnlyForMaster() && !onMasterNode {
			continue
		}
		if err := c.validateComponent(); err != nil {
			return err
		}
		if err := c.configureComponent(); err != nil {
			return err
		}
		if err := c.waitForComponentReady(); err != nil {
			return err
		}
	}
	return nil
}
