package kwok

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/kwok.yaml
var kw []byte

//go:embed manifests/stage-fast.yaml
var sf []byte

type KwokOpt struct {
	client k8s.K8sDynamicClient
}

func NewKwokOpt(client k8s.K8sDynamicClient) *KwokOpt {
	return &KwokOpt{
		client: client,
	}
}

func (k *KwokOpt) Exec() error {
	for _, yamlData := range [][]byte{kw, sf} {
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
			if err := k.client.Apply(obj); err != nil {
				return fmt.Errorf("error applying manifest %s", err)
			}
		}
	}
	return nil
}
