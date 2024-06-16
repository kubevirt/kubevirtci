package istio

import (
	"embed"
	"fmt"
	"log"
	"time"

	istiov1alpha1 "istio.io/operator/pkg/apis/istio/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

//go:embed manifests/*
var f embed.FS

type IstioOpt struct {
	sshPort     uint16
	cnaoEnabled bool
	client      k8s.K8sDynamicClient
}

func NewIstioOpt(c k8s.K8sDynamicClient, sshPort uint16, cnaoEnabled bool) *IstioOpt {
	return &IstioOpt{
		client:      c,
		sshPort:     sshPort,
		cnaoEnabled: cnaoEnabled,
	}
}

func (o *IstioOpt) Exec() error {
	istioCnao, err := f.ReadFile("manifests/istio-operator-with-cnao.yaml")
	if err != nil {
		return err
	}
	istioWithoutCnao, err := f.ReadFile("manifests/istio-operator-with-cnao.yaml")
	if err != nil {
		return err
	}
	err = o.client.Apply(f, "manifests/ns.yaml")
	if err != nil {
		return err
	}

	cmds := []string{
		"istioctl --kubeconfig /etc/kubernetes/admin.conf --hub quay.io/kubevirtci operator init",
		fmt.Sprintf("cat <<EOF > /opt/istio/istio-operator-with-cnao.cr.yaml\n%s\nEOF", string(istioCnao)),
		fmt.Sprintf("cat <<EOF > /opt/istio/istio-operator.cr.yaml\n%s\nEOF", string(istioWithoutCnao)),
	}
	for _, cmd := range cmds {
		_, err := utils.JumpSSH(o.sshPort, 1, cmd, true, true)
		if err != nil {
			return err
		}
	}
	confFile := "/opt/istio/istio-operator.cr.yaml"
	if o.cnaoEnabled {
		confFile = "/opt/istio/istio-operator-with-cnao.cr.yaml"
	}

	err = o.client.Apply(f, confFile)
	if err != nil {
		return err
	}

	operator := &istiov1alpha1.IstioOperator{}
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		obj, err := o.client.Get(schema.GroupVersionKind{Group: "install.istio.io",
			Version: "v1alpha1",
			Kind:    "IstioOperator"}, "istio-operator", "istio-system")

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, operator)
		if err != nil {
			return err
		}
		if operator.Status.Status == 3 {
			break
		}
		log.Println("Istio operator didn't move to Healthy status, sleeping for 5 seconds")
		time.Sleep(time.Second * 5)
	}
	if operator.Status.Status != 3 {
		return fmt.Errorf("Istio operator failed to move to Healthy status after max retries")
	}

	return nil
}
