package tests

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ = Describe("NFS Functional test", func() {
	It("should execute nfs-csi successfully", func() {
		pvc := &corev1.PersistentVolumeClaim{}

		operation := func() error {
			obj, err := k8sClient.Get(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "PersistentVolumeClaim",
			}, "pvc-nfs-dynamic", "nfs-csi")

			if err != nil {
				return err
			}

			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, pvc)
			if err != nil {
				return err
			}

			if pvc.Status.Phase != "Bound" {
				return fmt.Errorf("PVC didn't move to Bound phase")
			}

			return nil
		}

		backoffStrategy := backoff.NewExponentialBackOff()
		backoffStrategy.InitialInterval = 10 * time.Second
		backoffStrategy.MaxElapsedTime = 3 * time.Minute

		err := backoff.Retry(operation, backoffStrategy)
		Expect(err).NotTo(HaveOccurred())
	})
})
