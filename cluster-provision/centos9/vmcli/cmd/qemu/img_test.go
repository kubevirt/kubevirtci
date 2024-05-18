package qemu_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"kubevirt.io/kubevirtci/cluster-provision/centos9/vmcli/cmd/qemu"
)

var _ = Describe("Qemu qemu-img wrapper", func() {
	Describe("Parsing the disk information", func() {
		When("the disk information is valid", func() {
			const sampleQemuImgInfoOutput = `
			{
				"children": [
				{
					"name": "file",
					"info": {
						"children": [
						],
						"virtual-size": 197120,
						"filename": "test",
						"format": "file",
						"actual-size": 200704,
						"format-specific": {
							"type": "file",
							"data": {
							}
						},
						"dirty-flag": false
					}
				}
				],
				"virtual-size": 1048576,
				"filename": "test",
				"cluster-size": 65536,
				"format": "qcow2",
				"actual-size": 200704,
				"format-specific": {
					"type": "qcow2",
					"data": {
						"compat": "1.1",
						"compression-type": "zlib",
						"lazy-refcounts": false,
						"refcount-bits": 16,
						"corrupt": false,
						"extended-l2": false
					}
				},
				"dirty-flag": false
			}
			`
			It("returns the correct disk virtual size", func() {
				diskInfo, err := qemu.ParseDiskInfo([]byte(sampleQemuImgInfoOutput))
				Expect(err).NotTo(HaveOccurred())
				Expect(diskInfo.VirtualSize).To(BeNumerically("==", 1048576))
			})
		})

		When("the disk information is invalid", func() {
			It("fails", func() {
				_, err := qemu.ParseDiskInfo([]byte("{"))
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
