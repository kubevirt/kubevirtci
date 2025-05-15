package rookceph

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed manifests/*
var f embed.FS

type cephOpt struct {
	client    k8s.K8sDynamicClient
	sshClient libssh.Client
}

func NewCephOpt(c k8s.K8sDynamicClient, sshClient libssh.Client) *cephOpt {
	return &cephOpt{
		client:    c,
		sshClient: sshClient,
	}
}

func (o *cephOpt) Exec() error {
	err := fs.WalkDir(f, "manifests", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".yaml" {
			yamlData, err := f.ReadFile(path)
			if err != nil {
				return err
			}
			yamlDocs := bytes.Split(yamlData, []byte("---\n"))
			for _, yamlDoc := range yamlDocs {
				if len(yamlDoc) == 0 {
					continue
				}

				obj, err := k8s.SerializeIntoObject(yamlDoc)
				if err != nil {
					logrus.WithField("path", path).Info(err.Error())
					continue
				}
				if err := o.client.Apply(obj); err != nil {
					return fmt.Errorf("error applying manifest %q: %v", path, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	blockpool := &cephv1.CephBlockPool{}

	operation := func() error {
		obj, err := o.client.Get(schema.GroupVersionKind{
			Group:   "ceph.rook.io",
			Version: "v1",
			Kind:    "CephBlockPool",
		}, "replicapool", "rook-ceph")

		if err != nil {
			logrus.Errorf("Attempt failed: %v", err)
			return err
		}

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, blockpool)
		if err != nil {
			logrus.Errorf("Attempt failed: %v", err)
			return err
		}

		if blockpool.Status == nil {
			err := fmt.Errorf("ceph block pool: no status yet")
			logrus.Info(err)
			return err
		}

		if blockpool.Status.Phase != "Ready" {
			err := fmt.Errorf("ceph block pool phase=%q: CephBlockPool=%+v", blockpool.Status.Phase, blockpool)
			logrus.Info(err)
			return err
		}

		logrus.Infof("ceph block pool phase=%q: CephBlockPool=%+v", blockpool.Status.Phase, blockpool)

		return nil
	}

	maxElapsedTime := 10 * time.Minute
	err = backoff.Retry(operation, backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(45*time.Second),
		backoff.WithMaxInterval(90*time.Second),
		backoff.WithMaxElapsedTime(maxElapsedTime),
	))
	if err != nil {
		return fmt.Errorf("operation timed out after %s: %w", maxElapsedTime, err)
	}

	cmds := []string{
		`kubectl --kubeconfig /etc/kubernetes/admin.conf patch storageclass local -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'`,
		`kubectl --kubeconfig /etc/kubernetes/admin.conf patch storageclass rook-ceph-block -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'`,
	}
	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}
