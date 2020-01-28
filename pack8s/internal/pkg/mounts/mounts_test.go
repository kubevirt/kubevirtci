package mounts_test

import (
	"bufio"
	"bytes"
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logger "github.com/apsdehal/go-logger"

	"github.com/fromanirh/pack8s/internal/pkg/ledger"
	"github.com/fromanirh/pack8s/internal/pkg/mounts"
	"github.com/fromanirh/pack8s/internal/pkg/podman"
	"github.com/fromanirh/pack8s/iopodman"
)

func NewLogger() *logger.Logger {
	log, err := logger.New("test", 0, logger.DebugLevel)
	if err != nil {
		panic(err)
	}
	return log
}

var _ = Describe("mounts", func() {
	ctx := context.Background()

	log := NewLogger()
	hnd, _ := podman.NewHandle(ctx, "", log)

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	ldgr := ledger.NewLedger(hnd, w, log)

	defer func() {
		ldgr.Done <- nil
	}()

	Context("mounts", func() {
		It("Should mount volume to container", func() {
			name := "pack8s-test"

			registryMounts, err := mounts.NewVolumeMappings(ldgr, []mounts.MountInfo{
				{
					Name: name,
					Path: "/tmp",
					Type: "volume",
				},
			})
			Expect(err).To(BeNil())

			volumes, err := hnd.GetPrefixedVolumes(name)
			Expect(err).To(BeNil())

			found := false
			for _, volume := range volumes {
				if volume.Name == name {
					found = true
				}
			}
			Expect(found).To(Equal(true))

			Expect(len(registryMounts.ToStrings())).To(Equal(1))
			Expect(registryMounts.ToStrings()[0]).To(Equal("type=volume,source=pack8s-test,destination=/tmp"))

			volumesToRemove := []iopodman.Volume{
				{
					Name: name,
				},
			}
			err = hnd.RemoveVolumes(volumesToRemove)
			Expect(err).To(BeNil())
		})
	})
})
