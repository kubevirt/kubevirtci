package tests

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("Ceph Functional test", func() {
	It("should execute ceph successfully", func() { // test for something else, the pvc wont exist
		blockpool := &cephv1.CephBlockPool{}

		operation := func() error {
			obj, err := k8sClient.Get(schema.GroupVersionKind{
				Group:   "ceph.rook.io",
				Version: "v1",
				Kind:    "CephBlockPool",
			}, "replicapool", "rook-ceph")

			if err != nil {
				return err
			}

			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, blockpool)
			if err != nil {
				return err
			}

			if blockpool.Status == nil || blockpool.Status.Phase != "Ready" {
				return fmt.Errorf("Ceph pool block didn't move to ready status")
			}

			return nil
		}

		backoffStrategy := backoff.NewExponentialBackOff()
		backoffStrategy.InitialInterval = 30 * time.Second
		backoffStrategy.MaxElapsedTime = 6 * time.Minute

		err := backoff.Retry(operation, backoffStrategy)
		Expect(err).NotTo(HaveOccurred())
	})
})
