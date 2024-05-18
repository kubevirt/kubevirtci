package utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"kubevirt.io/kubevirtci/cluster-provision/centos9/vmcli/cmd/utils"
)

var _ = Describe("VM disk utils", func() {
	var fs afero.Fs
	var diskUtil *utils.DiskUtil

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		diskUtil = utils.NewDiskUtil(fs)
	})

	Describe("Calculating the next disk path", func() {
		BeforeEach(func() {
			createMockFile(fs, "/random.qcow2")
		})

		DescribeTable("When there are no already-existing disks",
			func(forcedDisk string, expectedNextFisk string, expectedCurrentDisk string) {
				nextDisk, currentDisk, err := diskUtil.CalcNextDisk("/", forcedDisk)
				Expect(err).NotTo(HaveOccurred())
				Expect(nextDisk).To(Equal(expectedNextFisk))
				Expect(currentDisk).To(Equal(expectedCurrentDisk))
			},
			Entry("returns the first disk as available and the default disk as the current one", "", "/disk01.qcow2", "box.qcow2"),
			Entry("returns the forced next disk name", "/forced.qcow2", "/forced.qcow2", "box.qcow2"),
		)

		DescribeTable("When there are disks already existing",
			func(forcedDisk string, expectedNextFisk string, expectedCurrentDisk string) {
				createMockFile(fs, "/disk01.qcow2")
				createMockFile(fs, "/disk02.qcow2")

				nextDisk, currentDisk, err := diskUtil.CalcNextDisk("/", forcedDisk)
				Expect(err).NotTo(HaveOccurred())
				Expect(nextDisk).To(Equal(expectedNextFisk))
				Expect(currentDisk).To(Equal(expectedCurrentDisk))
			},
			Entry("returns the next disk available and the last disk as the current one", "", "/disk03.qcow2", "/disk02.qcow2"),
			Entry("returns the forced next disk name", "/forced.qcow2", "/forced.qcow2", "/disk02.qcow2"),
		)
	})
})
