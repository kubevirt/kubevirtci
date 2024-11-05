package controlplane

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed config/token.yaml
var token []byte

//go:embed config/kubelet-kubeadm.yaml
var kk []byte

//go:embed config/cluster-info.yaml
var clusterInfo []byte

type BootstrapAuthResourcesPhase struct {
	client  k8s.K8sDynamicClient
	pkiPath string
}

func NewBootstrapAuthResourcesPhase(client k8s.K8sDynamicClient, pkiPath string) *BootstrapAuthResourcesPhase {
	return &BootstrapAuthResourcesPhase{
		client:  client,
		pkiPath: pkiPath,
	}
}

func (p *BootstrapAuthResourcesPhase) Run() error {
	caData, err := os.ReadFile(path.Join(p.pkiPath + "ca.crt"))
	if err != nil {
		return err
	}
	clusterInfo = []byte(strings.Replace(string(clusterInfo), "CADATA", base64.StdEncoding.EncodeToString(caData), 1))
	for _, manifest := range [][]byte{token, kk, clusterInfo} {
		yamlDocs := bytes.Split(manifest, []byte("---\n"))
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
	}
	return nil
}
