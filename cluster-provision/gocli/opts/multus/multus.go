package multus

import (
	"embed"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

//go:embed manifests/*
var f embed.FS

type MultusOpt struct {
	client k8s.K8sDynamicClient
}

func NewMultusOpt(c k8s.K8sDynamicClient) *MultusOpt {
	return &MultusOpt{
		client: c,
	}
}

func (o *MultusOpt) Exec() error {
	yamlData, err := f.ReadFile("manifests/multus.yaml")
	if err != nil {
		return err
	}
	if err := o.client.Apply(yamlData); err != nil {
		return err
	}
	return nil
}
