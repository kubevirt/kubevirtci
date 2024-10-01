package controlplane

import "kubevirt.io/kubevirtci/cluster-provision/gocli/cri"

const (
	etcdImage         = "etcd:3.5.10-0"
	apiServer         = "kube-apiserver"
	controllerManager = "kube-controller-manager"
	scheduler         = "kube-scheduler"
	registry          = "registry.k8s.io"
)

type ControlPlane interface{}

type ControlPlaneRunner struct {
	Phases []Phase
}

type Phase interface {
	Run() error
}

func NewControlPlaneRunner(containerRuntime cri.ContainerClient) *ControlPlaneRunner {
	phases := []Phase{}

	phases = append(phases, NewRunETCDPhase("", containerRuntime))
	phases = append(phases, NewRunControlPlaneComponentsPhase("", containerRuntime))

	return &ControlPlaneRunner{
		Phases: phases,
	}
}

func (cp *ControlPlaneRunner) Start() error {
	for _, phase := range cp.Phases {
		err := phase.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
