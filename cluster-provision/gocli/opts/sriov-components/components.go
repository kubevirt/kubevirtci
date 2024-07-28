package sriovcomponents

import (
	"embed"
	"fmt"

	"bytes"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/*
var f embed.FS

type sriovComponentsOpt struct {
	client k8s.K8sDynamicClient
}

func NewSriovComponentsOpt(c k8s.K8sDynamicClient) *sriovComponentsOpt {
	return &sriovComponentsOpt{
		client: c,
	}
}

func (o *sriovComponentsOpt) Exec() error {
	yamlData, err := f.ReadFile("manifests/components.yaml")
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
