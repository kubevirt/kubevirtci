package rookceph

import (
	"embed"
	"fmt"
	"time"

	cephv1 "github.com/aerosouund/rook/pkg/apis/ceph.rook.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
)

//go:embed manifests/*
var f embed.FS

type CephOpt struct {
	client k8s.K8sDynamicClient
}

func NewCephOpt(c k8s.K8sDynamicClient) *CephOpt {
	return &CephOpt{
		client: c,
	}
}

func (o *CephOpt) Exec() error {
	manifests := []string{
		"manifests/snapshot.storage.k8s.io_volumesnapshots.yaml",
		"manifests/snapshot.storage.k8s.io_volumesnapshotcontents.yaml",
		"manifests/snapshot.storage.k8s.io_volumesnapshotclasses.yaml",
		"manifests/rbac-snapshot-controller.yaml",
		"manifests/setup-snapshot-controller.yaml",
		"manifests/common.yaml",
		"manifests/crds.yaml",
		"manifests/operator.yaml",
		"manifests/cluster-test.yaml",
		"manifests/pool-test.yaml",
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

	blockpool := &cephv1.CephBlockPool{}
	maxRetries := 12

	for i := 0; i < maxRetries; i++ {
		obj, err := o.client.Get(schema.GroupVersionKind{
			Group:   "ceph.rook.io",
			Version: "v1",
			Kind:    "CephBlockPool"},
			"replicapool",
			"rook-ceph")

		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, blockpool)
		if err != nil {
			return err
		}

		if blockpool.Status != nil && blockpool.Status.Phase == "Ready" {
			break
		}
		fmt.Println("Ceph pool block didn't move to ready status, sleeping for 10 seconds")
		time.Sleep(10 * time.Second)
	}

	if blockpool.Status != nil && blockpool.Status.Phase != "Ready" {
		return fmt.Errorf("CephBlockPool replica pool did not become ready after %d retries", maxRetries)
	}

	return nil
}
