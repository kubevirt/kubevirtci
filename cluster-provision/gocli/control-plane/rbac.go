package controlplane

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed config/rbac.yaml
var rbac []byte

type CreateBootstrappersRBACPhase struct {
	client k8s.K8sDynamicClient
}

func NewCreateBootstrappersRBACPhase(client k8s.K8sDynamicClient) *CreateBootstrappersRBACPhase {
	return &CreateBootstrappersRBACPhase{
		client: client,
	}
}

func (p *CreateBootstrappersRBACPhase) Run() error {
	yamlDocs := bytes.Split(rbac, []byte("---\n"))
	for _, yamlDoc := range yamlDocs {
		if len(yamlDoc) == 0 {
			continue
		}

		obj, err := k8s.SerializeIntoObject(yamlDoc)
		if err != nil {
			logrus.Info(err.Error())
			continue
		}
		if err := p.client.Apply(obj); err != nil {
			return fmt.Errorf("error applying manifest %s", err)
		}
	}
	return nil
}
