package istio

import (
	"embed"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
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
	version     string
	sshClient   utils.SSHClient
}

func NewIstioOpt(sc utils.SSHClient, c k8s.K8sDynamicClient, sshPort uint16, cnaoEnabled bool) *IstioOpt {
	return &IstioOpt{
		client:      c,
		sshPort:     sshPort,
		cnaoEnabled: cnaoEnabled,
		version:     "1.15.0",
		sshClient:   sc,
	}
}

func (o *IstioOpt) Exec() error {
	yamlData, err := f.ReadFile("manifests/ns.yaml")
	if err != nil {
		return err
	}
	if err := o.client.Apply(yamlData); err != nil {
		return err
	}

	cmds := []string{
		"source /var/lib/kubevirtci/shared_vars.sh",
		"PATH=/opt/istio-" + o.version + "/bin:$PATH istioctl --kubeconfig /etc/kubernetes/admin.conf --hub quay.io/kubevirtci operator init",
	}
	for _, cmd := range cmds {
		if _, err := o.sshClient.JumpSSH(o.sshPort, 1, cmd, true, true); err != nil {
			return err
		}
	}

	if o.cnaoEnabled {
		yamlData, err := f.ReadFile("manifests/istio-operator-with-cnao.cr.yaml")
		if err != nil {
			return err
		}
		if err := o.client.Apply(yamlData); err != nil {
			return err
		}
	} else {
		yamlData, err := f.ReadFile("manifests/istio-operator.cr.yaml")
		if err != nil {
			return err
		}
		if err := o.client.Apply(yamlData); err != nil {
			return err
		}
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
		if operator.Status != nil && operator.Status.Status == 3 {
			break
		}
		logrus.Info("Istio operator didn't move to Healthy status, sleeping for 10 seconds")
		time.Sleep(time.Second * 10)
	}
	if operator.Status.Status != 3 {
		return fmt.Errorf("Istio operator failed to move to Healthy status after max retries")
	}
	logrus.Info("Istio operator is now ready!")

	return nil
}
