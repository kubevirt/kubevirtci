package cnao

import (
	"embed"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

//go:embed manifests/*
var f embed.FS

type CnaoOpt struct {
	client k8s.K8sDynamicClient
}

func NewCnaoOpt(c k8s.K8sDynamicClient) *CnaoOpt {
	return &CnaoOpt{
		client: c,
	}
}

func (o *CnaoOpt) Exec() error {
	manifests := []string{
		"manifests/ns.yaml",
		"manifests/crd.yaml",
		"manifests/operator.yaml",
	}
	for _, manifest := range manifests {
		if err := o.client.Apply(f, manifest); err != nil {
			return err
		}
	}
	return nil
}
