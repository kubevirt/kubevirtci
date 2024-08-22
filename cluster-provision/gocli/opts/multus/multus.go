package multus

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/multus.yaml
var multus []byte

type multusOpt struct {
	client    k8s.K8sDynamicClient
	sshClient libssh.Client
}

func NewMultusOpt(c k8s.K8sDynamicClient, sshClient libssh.Client) *multusOpt {
	return &multusOpt{
		client:    c,
		sshClient: sshClient,
	}
}

func (o *multusOpt) Exec() error {
	yamlDocs := bytes.Split(multus, []byte("---\n"))
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
	if err := o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf rollout status -n kube-system ds/kube-multus-ds --timeout=200s"); err != nil {
		return err
	}
	return nil
}
