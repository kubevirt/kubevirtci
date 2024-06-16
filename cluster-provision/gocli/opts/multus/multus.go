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
	if err := o.client.Apply(f, "manifests/multus.yaml"); err != nil {
		return err
	}
	return nil
}
