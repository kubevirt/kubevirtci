package multus

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"time"

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
	var rolloutComplete bool
	go func() {
		err := o.sshClient.Command("kubectl --kubeconfig=/etc/kubernetes/admin.conf rollout status -n kube-system ds/kube-multus-ds --timeout=200s")
		if err != nil {
			fmt.Println("Rollout status failed:", err)
		}
		rolloutComplete = true
	}()

	for {
		if rolloutComplete {
			break
		}
		cmd := "kubectl --kubeconfig=/etc/kubernetes/admin.conf get pods -n kube-system -l name=kube-multus-ds -o jsonpath='{.items[*].metadata.name}'"
		pods, err := o.sshClient.CommandWithNoStdOut(cmd)
		if err != nil {
			fmt.Println("Failed to get pods:", err)
			return err
		}

		fmt.Println("Pods in kube-multus-ds DaemonSet:", pods)
		for _, pod := range strings.Split(pods, " ") {
			logCmd := fmt.Sprintf("kubectl --kubeconfig=/etc/kubernetes/admin.conf logs -n kube-system %s --tail=30", pod)
			err := o.sshClient.Command(logCmd)
			if err != nil {
				fmt.Println("Failed to get logs for pod", pod, ":", err)
				continue
			}
		}
		time.Sleep(10 * time.Second)
	}
	return nil
}
