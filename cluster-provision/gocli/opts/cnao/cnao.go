package cnao

import (
	"embed"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/*
var f embed.FS

type CnaoOpt struct {
	client    k8s.K8sDynamicClient
	sshClient libssh.Client
}

func NewCnaoOpt(c k8s.K8sDynamicClient, sshClient libssh.Client) *CnaoOpt {
	return &CnaoOpt{
		client:    c,
		sshClient: sshClient,
	}
}

func (o *CnaoOpt) Exec() error {
	manifests := []string{
		"manifests/ns.yaml",
		"manifests/crd.yaml",
		"manifests/operator.yaml",
		"manifests/whereabouts.yaml",
	}
	for _, manifest := range manifests {
		yamlData, err := f.ReadFile(manifest)
		if err != nil {
			return err
		}
		if err := o.client.Apply(yamlData); err != nil {
			return err
		}
	}

	if _, err := o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf wait deployment -n cluster-network-addons cluster-network-addons-operator --for condition=Available --timeout=200s", true); err != nil {
		return err
	}
	return nil
}
