package istio

import (
	_ "embed"

	"github.com/sirupsen/logrus"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/ns.yaml
var ns []byte

//go:embed manifests/istio-operator-with-cnao.cr.yaml
var istioWithCnao []byte

//go:embed manifests/istio-operator.cr.yaml
var istioNoCnao []byte

const istioVersion = "1.23.1"

type istioOpt struct {
	cnaoEnabled bool
	client      k8s.K8sDynamicClient
	sshClient   libssh.Client
}

func NewIstioOpt(sc libssh.Client, c k8s.K8sDynamicClient, cnaoEnabled bool) *istioOpt {
	return &istioOpt{
		client:      c,
		cnaoEnabled: cnaoEnabled,
		sshClient:   sc,
	}
}

func (o *istioOpt) Exec() error {
	obj, err := k8s.SerializeIntoObject(ns)
	if err != nil {
		return err
	}

	if err := o.client.Apply(obj); err != nil {
		return err
	}

	istioFile := "/opt/istio-operator-with-cnao.yaml"
	if !o.cnaoEnabled {
		istioFile = "/opt/istio-operator.cr.yaml"
	}

	cmds := []string{
		"source /var/lib/kubevirtci/shared_vars.sh",
		`echo '` + string(istioWithCnao) + `' |  tee /opt/istio-operator-with-cnao.yaml > /dev/null`,
		`echo '` + string(istioNoCnao) + `' |  tee /opt/istio-operator.cr.yaml > /dev/null`,
	}
	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	istioInstallCmd := "PATH=/opt/istio-" + istioVersion + "/bin:$PATH istioctl --kubeconfig /etc/kubernetes/admin.conf install -y -f " + istioFile
	if err := o.sshClient.Command(istioInstallCmd); err != nil {
		return err
	}

	logrus.Info("Istio operator is now ready!")
	return nil
}
