package nfscsi

import (
	"embed"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
)

//go:embed manifests/*
var f embed.FS

type NfsCsiOpt struct {
	client k8s.K8sDynamicClient
}

func NewNfsCsiOpt(c k8s.K8sDynamicClient) *NfsCsiOpt {
	return &NfsCsiOpt{
		client: c,
	}
}

func (o *NfsCsiOpt) Exec() error {
	manifests := []string{
		"manifests/nfs-service.yaml",
		"manifests/nfs-server.yaml",
		"manifests/csi-nfs-controller-rbac.yaml",
		"manifests/csi-nfs-driverinfo.yaml",
		"manifests/csi-nfs-controller.yaml",
		"manifests/csi-nfs-node.yaml",
		"manifests/csi-nfs-sc.yaml",
		"manifests/csi-nfs-test-pvc.yaml",
	}

	for _, manifest := range manifests {
		yamlData, err := f.ReadFile(manifest)
		if err != nil {
			return err
		}
		if err := o.client.Apply(yamlData); err != nil {
			return err
		}
	}
	pvc := &corev1.PersistentVolumeClaim{}
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		obj, err := o.client.Get(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "PersistentVolumeClaim"},
			"pvc-nfs-dynamic",
			"nfs-csi")
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, pvc)
		if err != nil {
			return err
		}

		if pvc.Status.Phase == "Bound" {
			break
		}
		logrus.Info("PVC didn't move to Bound phase, sleeping for 10 seconds")
		time.Sleep(10 * time.Second)
	}

	if pvc.Status.Phase != "Bound" {
		return fmt.Errorf("PVC failed to transition to Bound!")
	}

	err := o.client.Delete(schema.GroupVersionKind{
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
