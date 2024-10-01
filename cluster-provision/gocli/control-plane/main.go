package controlplane

import "kubevirt.io/kubevirtci/cluster-provision/gocli/cri"

const (
	etcdImage         = "etcd:3.5.10-0"
	apiServer         = "kube-apiserver"
	controllerManager = "kube-controller-manager"
	scheduler         = "kube-scheduler"
	registry          = "registry.k8s.io"
)

type ControlPlaneRunner struct {
	dnsmasqID        string
	containerRuntime cri.ContainerClient
}

type Phase interface {
	Run() error
}

func NewControlPlaneRunner() {}

func (cp *ControlPlaneRunner) Start() error {
	if err := NewRunETCDPhase(cp.dnsmasqID, cp.containerRuntime).Run(); err != nil {
		return err
	}

	if err := NewRunControlPlaneComponentsPhase(cp.dnsmasqID, cp.containerRuntime).Run(); err != nil {
		return err
	}

	return nil
}
