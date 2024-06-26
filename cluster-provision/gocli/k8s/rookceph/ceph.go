package rookceph

import (
	"fmt"
	"time"

	cephv1 "github.com/aerosouund/rook/pkg/apis/ceph.rook.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/k8s/common"
)

type CephOpt struct {
	client *k8s.K8sDynamicClient
}

func NewCephOpt(c *k8s.K8sDynamicClient) *CephOpt {
	return &CephOpt{
		client: c,
	}
}

func (o *CephOpt) Exec() error {
	manifests := []string{
		"/workdir/manifests/ceph/snapshot.storage.k8s.io_volumesnapshots.yaml",
		"/workdir/manifests/ceph/snapshot.storage.k8s.io_volumesnapshotcontents.yaml",
		"/workdir/manifests/ceph/snapshot.storage.k8s.io_volumesnapshotclasses.yaml",
		"/workdir/manifests/ceph/rbac-snapshot-controller.yaml",
		"/workdir/manifests/ceph/setup-snapshot-controller.yaml",
		"/workdir/manifests/ceph/common.yaml",
		"/workdir/manifests/ceph/crds.yaml",
		"/workdir/manifests/ceph/operator.yaml",
		"/workdir/manifests/ceph/cluster-test.yaml",
		"/workdir/manifests/ceph/pool-test.yaml",
	}

	for _, manifest := range manifests {
		err := o.client.Apply(manifest)
		if err != nil {
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

	// if blockpool.Status != nil && blockpool.Status.Phase != "Ready" {
	// 	return fmt.Errorf("CephBlockPool replica pool did not become ready after %d retries", maxRetries)
	// }

	return nil
}
