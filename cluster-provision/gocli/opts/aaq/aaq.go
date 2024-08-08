package aaq

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/operator.yaml
var operator []byte

//go:embed manifests/cr.yaml
var cr []byte

type aaqOpt struct {
	client        k8s.K8sDynamicClient
	sshClient     libssh.Client
	customVersion string
}

func NewAaqOpt(c k8s.K8sDynamicClient, sshClient libssh.Client, customVersion string) *aaqOpt {
	return &aaqOpt{
		client:        c,
		sshClient:     sshClient,
		customVersion: customVersion,
	}
}

func (o *aaqOpt) Exec() error {
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

	if err := o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait --for=condition=Ready pod --timeout=180s --all --namespace aaq"); err != nil {
		return err
	}
	logrus.Info("AAQ Operator is ready!")
	return nil
}
