package ledger_test

import (
	"bufio"
	"bytes"
	"context"

	logger "github.com/apsdehal/go-logger"

	"github.com/fromanirh/pack8s/internal/pkg/images"
	"github.com/fromanirh/pack8s/internal/pkg/ledger"
	"github.com/fromanirh/pack8s/internal/pkg/podman"
	"github.com/fromanirh/pack8s/iopodman"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func NewLogger() *logger.Logger {
	log, err := logger.New("test", 0, logger.DebugLevel)
	if err != nil {
		panic(err)
	}
	return log
}

var _ = Describe("ledger", func() {
	ctx := context.Background()

	log := NewLogger()
	hnd, _ := podman.NewHandle(ctx, "", log)

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	ldgr := ledger.NewLedger(hnd, w, log)

	defer func() {
		ldgr.Done <- nil
	}()

	Context("create new volume", func() {
		It("Should create new volume without any error", func() {
			volume, err := ldgr.MakeVolume("pack8s-test")
			Expect(err).To(BeNil())
			Expect(volume).NotTo(Equal(""))

			volumes, err := hnd.GetAllVolumes()
			Expect(err).To(BeNil())

			found := false
			for _, volume := range volumes {
				if volume.Name == "pack8s-test" {
					found = true
				}
			}

			Expect(found).To(Equal(true))

			volumesToRemove := []iopodman.Volume{
				{
					Name: "pack8s-test",
				},
			}
			err = hnd.RemoveVolumes(volumesToRemove)
			Expect(err).To(BeNil())
		})

	})

	Context("create and run new container", func() {
		It("Should create and run new container without any error", func() {
			name := "pack8s-test"
			id, err := ldgr.RunContainer(iopodman.Create{
				Args: []string{images.DockerRegistryImage},
				Name: &name,
			})
			Expect(err).To(BeNil())
			Expect(id).NotTo(Equal(""))

			_, err = hnd.StopContainer(id, 10)
			Expect(err).To(BeNil())

			_, err = hnd.RemoveContainer(iopodman.Container{Id: id}, true, true)
			Expect(err).To(BeNil())
		})

	})

})
