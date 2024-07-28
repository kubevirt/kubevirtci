package multussriov

import (
	"embed"
	"fmt"

	"bytes"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/*
var f embed.FS

type multusSriovOpt struct {
	client k8s.K8sDynamicClient
}

func NewMultusSriovOpt(c k8s.K8sDynamicClient) *multusSriovOpt {
	return &multusSriovOpt{
		client: c,
	}
}

func (o *multusSriovOpt) Exec() error {
	yamlData, err := f.ReadFile("manifests/multus.yaml")
	if err != nil {
		return err
	}
	yamlDocs := bytes.Split(yamlData, []byte("---\n"))
	for _, yamlDoc := range yamlDocs {
		if len(yamlDoc) == 0 {
			continue
		}

		obj, err := k8s.SerializeIntoObject(yamlDoc)
		if err != nil {
			logrus.Info(err.Error())
			continue
		}
		if err := o.client.Apply(obj); err != nil {
			return fmt.Errorf("error applying manifest %s", err)
		}
	}
	return nil
}
