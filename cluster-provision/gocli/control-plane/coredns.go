package controlplane

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed config/coredns.yaml
var coredns []byte

type CoreDNSPhase struct {
	client k8s.K8sDynamicClient
}

func NewCoreDNSPhase(client k8s.K8sDynamicClient) *CoreDNSPhase {
	return &CoreDNSPhase{
		client: client,
	}
}

func (c *CoreDNSPhase) Run() error {
	yamlDocs := bytes.Split(coredns, []byte("---\n"))
	for _, yamlDoc := range yamlDocs {
		if len(yamlDoc) == 0 {
			continue
		}

		obj, err := k8s.SerializeIntoObject(yamlDoc)
		if err != nil {
			logrus.Info(err.Error())
			continue
		}
		if err := c.client.Apply(obj); err != nil {
			return fmt.Errorf("error applying manifest %s", err)
		}
	}
	return nil
}
