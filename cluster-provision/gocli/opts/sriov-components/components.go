package sriovcomponents

import (
	"embed"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/*
var f embed.FS

type SriovComponentsOpt struct {
	client k8s.K8sDynamicClient
}

func NewSriovComponentsOpt(c k8s.K8sDynamicClient) *SriovComponentsOpt {
	return &SriovComponentsOpt{
		client: c,
	}
}

func (o *SriovComponentsOpt) Exec() error {
	yamlData, err := f.ReadFile("manifests/components.yaml")
	if err != nil {
		return err
	}
	if err := o.client.Apply(yamlData); err != nil {
		return err
	}
	return nil
}
