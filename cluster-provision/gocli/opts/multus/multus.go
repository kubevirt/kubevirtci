package multus

import (
	"embed"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/*
var f embed.FS

type MultusOpt struct {
	client    k8s.K8sDynamicClient
	sshClient libssh.Client
}

func NewMultusOpt(c k8s.K8sDynamicClient, sshClient libssh.Client) *MultusOpt {
	return &MultusOpt{
		client:    c,
		sshClient: sshClient,
	}
}

func (o *MultusOpt) Exec() error {
	yamlData, err := f.ReadFile("manifests/multus.yaml")
	if err != nil {
		return err
	}
	if err := o.client.Apply(yamlData); err != nil {
		return err
	}

	if _, err = o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf rollout status -n kube-system ds/kube-multus-ds --timeout=200s", true); err != nil {
		return err
	}
	return nil
}
