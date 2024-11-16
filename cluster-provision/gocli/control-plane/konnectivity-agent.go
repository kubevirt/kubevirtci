package controlplane

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed config/konnectivity-agent.yaml
var ka []byte

type KonnectivityAgentPhase struct {
	client k8s.K8sDynamicClient
}

func NewKonnectivityAgentPhase(client k8s.K8sDynamicClient) *KonnectivityAgentPhase {
	return &KonnectivityAgentPhase{
		client: client,
	}
}

func (k *KonnectivityAgentPhase) Run() error {
	yamlDocs := bytes.Split(kp, []byte("---\n"))
	for _, yamlDoc := range yamlDocs {
		if len(yamlDoc) == 0 {
			continue
		}

		obj, err := k8s.SerializeIntoObject(yamlDoc)
		if err != nil {
			logrus.Info(err.Error())
			continue
		}
		if err := k.client.Apply(obj); err != nil {
			return fmt.Errorf("error applying manifest %s", err)
		}
	}
	return nil
}
