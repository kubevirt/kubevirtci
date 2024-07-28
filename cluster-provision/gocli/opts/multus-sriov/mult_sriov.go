package multussriov

import (
	"embed"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/*
var f embed.FS

type MultusSriovOpt struct {
	client k8s.K8sDynamicClient
}

func NewMultusSriovOpt(c k8s.K8sDynamicClient) *MultusSriovOpt {
	return &MultusSriovOpt{
		client: c,
	}
}

func (o *MultusSriovOpt) Exec() error {
	yamlData, err := f.ReadFile("manifests/multus.yaml")
	if err != nil {
		return err
	}
	if err := o.client.Apply(yamlData); err != nil {
		return err
	}
	return nil
}
