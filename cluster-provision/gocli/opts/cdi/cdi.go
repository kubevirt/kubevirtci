package cdi

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/cdi-operator.yaml
var operator []byte

//go:embed manifests/cdi-cr.yaml
var cr []byte

type cdiOpt struct {
	client        k8s.K8sDynamicClient
	sshClient     libssh.Client
	customVersion string
}

func NewCdiOpt(c k8s.K8sDynamicClient, sshClient libssh.Client, cv string) *cdiOpt {
	return &cdiOpt{
		client:        c,
		sshClient:     sshClient,
		customVersion: cv,
	}
}

func (o *cdiOpt) Exec() error {
	if o.customVersion != "" {
		pattern := `v[0-9]+\.[0-9]+\.[0-9]+(.*)?$`
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		operatorNewVersion := regex.ReplaceAllString(string(operator), o.customVersion)
		operator = []byte(operatorNewVersion)
	}

	for _, manifest := range [][]byte{operator, cr} {
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
			if err := o.client.Apply(obj); err != nil {
				return fmt.Errorf("error applying manifest %s", err)
			}
		}
	}

	return nil
}
