package network_resources_injector

import (
	_ "embed"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/common"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/auth.yaml
var auth []byte

//go:embed manifests/service.yaml
var service []byte

//go:embed manifests/server.yaml
var server []byte

type networkResourcesInjectorOpt struct {
	client    k8s.K8sDynamicClient
	sshClient libssh.Client
}

func NewNetworkResourcesInjectorOpt(sc libssh.Client, c k8s.K8sDynamicClient) *networkResourcesInjectorOpt {
	return &networkResourcesInjectorOpt{
		client:    c,
		sshClient: sc,
	}
}

func (o *networkResourcesInjectorOpt) Exec() error {
	if err := common.ApplyYAML(auth, o.client); err != nil {
		return err
	}
	obj, err := k8s.SerializeIntoObject(service)
	if err != nil {
		return err
	}

	if err = o.client.Apply(obj); err != nil {
		return err
	}

	if err = common.ApplyYAML(server, o.client); err != nil {
		return err
	}

	return o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf rollout status -n kube-system deploy/network-resources-injector --timeout=200s")
}
