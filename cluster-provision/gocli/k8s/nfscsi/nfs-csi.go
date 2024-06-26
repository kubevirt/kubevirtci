package nfscsi

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/k8s/common"
)

type NfsCsiOpt struct {
	client *k8s.K8sDynamicClient
}

func NewNfsCsiOpt(c *k8s.K8sDynamicClient) *NfsCsiOpt {
	return &NfsCsiOpt{
		client: c,
	}
}

func (o *NfsCsiOpt) Exec() error {
	manifests := []string{
		"/workdir/manifests/nfs-csi/nfs-service.yaml",
		"/workdir/manifests/nfs-csi/nfs-server.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-controller-rbac.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-driverinfo.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-controller.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-node.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-sc.yaml",
		"/workdir/manifests/nfs-csi/csi-nfs-test-pvc.yaml",
	}

	for _, manifest := range manifests {
		err := o.client.Apply(manifest)
		if err != nil {
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
		fmt.Println("PVC didn't move to Bound phase, sleeping for 10 seconds")
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

	return nil
}
