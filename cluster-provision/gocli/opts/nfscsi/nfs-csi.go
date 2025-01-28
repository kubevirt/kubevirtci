package nfscsi

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/*
var f embed.FS

type nfsCsiOpt struct {
	client k8s.K8sDynamicClient
}

func NewNfsCsiOpt(c k8s.K8sDynamicClient) *nfsCsiOpt {
	return &nfsCsiOpt{
		client: c,
	}
}

func (o *nfsCsiOpt) Exec() error {
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
					logrus.Info(err.Error())
					continue
				}
				if err := o.client.Apply(obj); err != nil {
					return fmt.Errorf("error applying manifest %s", err)
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	pvc := &corev1.PersistentVolumeClaim{}

	operation := func() error {
		obj, err := o.client.Get(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "PersistentVolumeClaim",
		}, "pvc-nfs-dynamic", "nfs-csi")

		if err != nil {
			logrus.Errorf("Attempt failed, PVC is still not bound: %v", err)
			return err
		}

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, pvc)
		if err != nil {
			logrus.Errorf("Attempt failed, PVC is still not bound: %v", err)
			return err
		}

		if pvc.Status.Phase != "Bound" {
			err := fmt.Errorf("PVC didn't move to Bound phase")
			logrus.Info(err)
			return err
		}

		return nil
	}

	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.InitialInterval = 10 * time.Second
	backoffStrategy.MaxElapsedTime = 5 * time.Minute

	err = backoff.Retry(operation, backoffStrategy)
	if err != nil {
		return fmt.Errorf("Waiting on PVC to become bound failed after maximum retries: %v", err)
	}

	err = o.client.Delete(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "PersistentVolumeClaim",
	}, "pvc-nfs-dynamic", "nfs-csi")
	if err != nil {
		return err
	}
	logrus.Info("NFS CSI installed successfully!")

	return nil
}
