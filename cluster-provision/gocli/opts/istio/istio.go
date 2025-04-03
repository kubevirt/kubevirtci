package istio

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/ns.yaml
var ns []byte

//go:embed manifests/istio-operator-with-cnao.cr.yaml
var istioWithCnao []byte

//go:embed manifests/istio-operator.cr.yaml
var istioNoCnao []byte

const istioVersion = "1.24.4"

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

	go func() {
		operation := func() error {
			obj, err := o.client.Get(schema.GroupVersionKind{Group: "apps",
				Version: "v1",
				Kind:    "DaemonSet"}, "istio-cni-node", "kube-system")
			if err != nil {
				fmt.Printf("Error getting the CNI DaemonSet: %s\n", err.Error())
				return err
			}

			cniDaemonSet := &appsv1.DaemonSet{}
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, cniDaemonSet)
			if err != nil {
				fmt.Printf("Error converting the CNI DaemonSet: %s\n", err.Error())
				return err
			}

			privileged := true
			cniDaemonSet.Spec.Template.Spec.Containers[0].SecurityContext.Privileged = &privileged
			newCniDaemonSet, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cniDaemonSet)
			if err != nil {
				fmt.Printf("Error converting the CNI DaemonSet: %s\n", err.Error())
				return err
			}

			err = o.client.Update(&unstructured.Unstructured{Object: newCniDaemonSet})
			if err != nil {
				fmt.Printf("Error patching the CNI DaemonSet: %s\n", err.Error())
				return err
			}
			return nil
		}

		backoffStrategy := backoff.NewExponentialBackOff()
		backoffStrategy.InitialInterval = 10 * time.Second
		backoffStrategy.MaxElapsedTime = 3 * time.Minute

		_ = backoff.Retry(operation, backoffStrategy)
	}()

	istioInstallCmd := "PATH=/opt/istio-" + istioVersion + "/bin:$PATH istioctl --kubeconfig /etc/kubernetes/admin.conf install -y -f " + istioFile
	if err := o.sshClient.Command(istioInstallCmd); err != nil {
		return err
	}

	logrus.Info("Istio operator is now ready!")
	return nil
}
