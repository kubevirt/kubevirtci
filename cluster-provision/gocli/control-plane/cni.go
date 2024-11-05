package controlplane

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed config/cni.yaml
var cni4 []byte

//go:embed config/cni_ipv6.yaml
var cni6 []byte

type CNIPhase struct {
	ipv6   bool
	client k8s.K8sDynamicClient
}

func NewCNIPhase(client k8s.K8sDynamicClient, ipv6 bool) *CNIPhase {
	return &CNIPhase{
		ipv6:   ipv6,
		client: client,
	}
}

func (c *CNIPhase) Run() error {
	cni := cni4
	if c.ipv6 {
		cni = cni6
	}

	yamlDocs := bytes.Split(cni, []byte("---\n"))
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
