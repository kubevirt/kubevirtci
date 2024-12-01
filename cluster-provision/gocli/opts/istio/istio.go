package istio

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
	istiov1alpha1 "istio.io/operator/pkg/apis/istio/v1alpha1"
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

const istioVersion = "1.24.1"

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

	cmds := []string{
		"source /var/lib/kubevirtci/shared_vars.sh",
		"PATH=/opt/istio-" + istioVersion + "/bin:$PATH istioctl --kubeconfig /etc/kubernetes/admin.conf install -y",
	}
	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	obj, err = k8s.SerializeIntoObject(istioWithCnao)
	if err != nil {
		return err
	}

	if o.cnaoEnabled {
		if err := o.client.Apply(obj); err != nil {
			return err
		}
	} else {
		obj, err = k8s.SerializeIntoObject(istioNoCnao)
		if err != nil {
			return err
		}

		if err := o.client.Apply(obj); err != nil {
			return err
		}
	}

	operator := &istiov1alpha1.IstioOperator{}

	operation := func() error {
		obj, err := o.client.Get(schema.GroupVersionKind{Group: "install.istio.io",
			Version: "v1alpha1",
			Kind:    "IstioOperator"}, "istio-operator", "istio-system")

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, operator)
		if err != nil {
			return err
		}

		if operator.Status == nil {
			err := fmt.Errorf("Operator status is still not ready")
			logrus.Info("Istio operator is still not ready, Backing off and retrying")
			return err
		}

		if operator.Status.Status != 3 {
			err := fmt.Errorf("Istio operator failed to move to Healthy status after max retries")
			logrus.Info("Istio operator is still not ready, Backing off and retrying")
			return err
		}

		return nil
	}

	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.InitialInterval = 10 * time.Second
	backoffStrategy.MaxElapsedTime = 3 * time.Minute

	err = backoff.Retry(operation, backoffStrategy)
	if err != nil {
		return fmt.Errorf("Waiting on istio operator to become ready failed after maximum retries: %v", err)
	}

	logrus.Info("Istio operator is now ready!")
	return nil
}
