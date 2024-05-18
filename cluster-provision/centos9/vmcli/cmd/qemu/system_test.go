package qemu_test

import (
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"kubevirt.io/kubevirtci/cluster-provision/centos9/vmcli/cmd/qemu"
)

var _ = Describe("Qemu qemu-system wrapper", func() {
	Describe("Generating the command line", func() {
		When("the system definition is valid", func() {
			var qemuSystem qemu.QemuSystem

			BeforeEach(func() {
				systemUuid, err := uuid.Parse("771a640e-9c3e-498e-9319-ed1a48519adb")
				Expect(err).NotTo(HaveOccurred())

				qemuSystem = qemu.QemuSystem{
					Arch:          "x86_64",
					Memory:        "2048M",
					CpuCount:      2,
					Numa:          1,
					KvmEnabled:    true,
					CpuModel:      "host,migratable=no,+invtsc",
					Machine:       "q35,accel=kvm,kernel_irqchip=split",
					SystemUuid:    systemUuid,
					VncServer:     ":01",
					SerialHostDev: "pty",
					InitrdPath:    "/initrd.img",
					KernelPath:    "/vmlinuz",
					KernelArgs: []string{
						"root=/dev/vda1 ro",
						"console=tty0",
					},
					Drives: []string{
						"format=qcow2,file=/disk1.qcow2,if=virtio,cache=unsafe",
						"format=qcow2,file=/disk2.qcow2,if=virtio,cache=unsafe",
					},
					Devices: []string{
						"virtio-rng-pci",
						"AC97",
					},
					Netdev: "tap,id=network0,ifname=tap01,script=no,downscript=no",
				}
			})

			It("generates the correct command line", func() {
				cmdline, err := qemuSystem.GenerateCmdline()
				Expect(err).NotTo(HaveOccurred())
				Expect(cmdline).To(Equal("qemu-system-x86_64 -m 2048M -smp 2 -cpu host,migratable=no,+invtsc -M q35,accel=kvm,kernel_irqchip=split -uuid 771a640e-9c3e-498e-9319-ed1a48519adb -vnc :01 -serial pty -initrd /initrd.img -kernel /vmlinuz -append \"root=/dev/vda1 ro console=tty0\" -netdev tap,id=network0,ifname=tap01,script=no,downscript=no -enable-kvm -drive format=qcow2,file=/disk1.qcow2,if=virtio,cache=unsafe -drive format=qcow2,file=/disk2.qcow2,if=virtio,cache=unsafe -device virtio-rng-pci -device AC97"))
			})
		})
	})

	Describe("Parsing the memory argument", func() {
		When("the argument is invalid", func() {
			var qemuSystem qemu.QemuSystem

			BeforeEach(func() {
				qemuSystem = qemu.QemuSystem{
					Memory: "invalid",
				}
			})

			It("fails", func() {
				_, _, err := qemuSystem.ParseMemory()
				Expect(err).To(HaveOccurred())
			})
		})

		When("there are no unit", func() {
			var qemuSystem qemu.QemuSystem

			BeforeEach(func() {
				qemuSystem = qemu.QemuSystem{
					Memory: "2048",
				}
			})

			It("returns the correct numeric value", func() {
				val, _, err := qemuSystem.ParseMemory()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(BeNumerically("==", 2048))
			})

			It("returns an empty unit", func() {
				_, unit, err := qemuSystem.ParseMemory()
				Expect(err).NotTo(HaveOccurred())
				Expect(unit).To(Equal(""))
			})
		})

		When("there is a unit", func() {
			var qemuSystem qemu.QemuSystem

			BeforeEach(func() {
				qemuSystem = qemu.QemuSystem{
					Memory: "2048M",
				}
			})

			It("returns the correct numeric value", func() {
				val, _, err := qemuSystem.ParseMemory()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(BeNumerically("==", 2048))
			})

			It("returns the correct unit", func() {
				_, unit, err := qemuSystem.ParseMemory()
				Expect(err).NotTo(HaveOccurred())
				Expect(unit).To(Equal("M"))
			})
		})
	})

	Describe("Generating the NUMA arguments", func() {
		When("there is a single NUMA node", func() {
			var qemuSystem qemu.QemuSystem

			BeforeEach(func() {
				qemuSystem = qemu.QemuSystem{
					Memory:   "3072M",
					CpuCount: 9,
					Numa:     1,
				}
			})

			It("generates no arguments", func() {
				numaArgs, err := qemuSystem.GenerateNumaArguments()
				Expect(err).NotTo(HaveOccurred())
				Expect(numaArgs).To(BeEmpty())
			})
		})

		When("there are multiple NUMA nodes", func() {
			var qemuSystem qemu.QemuSystem

			BeforeEach(func() {
				qemuSystem = qemu.QemuSystem{
					Memory:   "3072M",
					CpuCount: 9,
					Numa:     3,
				}
			})

			It("generates the correct arguments", func() {
				numaArgs, err := qemuSystem.GenerateNumaArguments()
				Expect(err).NotTo(HaveOccurred())
				Expect(numaArgs).To(ConsistOf(
					"-object memory-backend-ram,size=1024M,id=m0",
					"-numa node,nodeid=0,memdev=m0,cpus=0-2",
					"-object memory-backend-ram,size=1024M,id=m1",
					"-numa node,nodeid=1,memdev=m1,cpus=3-5",
					"-object memory-backend-ram,size=1024M,id=m2",
					"-numa node,nodeid=2,memdev=m2,cpus=6-8",
				))
			})
		})

		When("the CPU count is invalid", func() {
			var qemuSystem qemu.QemuSystem

			BeforeEach(func() {
				qemuSystem = qemu.QemuSystem{
					Memory:   "3072M",
					CpuCount: 8,
					Numa:     3,
				}
			})

			It("fails", func() {
				_, err := qemuSystem.GenerateNumaArguments()
				Expect(err).To(MatchError("unable to calculate symmetric NUMA topology with vCPUs:8 Memory:3072M NUMA:3"))
			})
		})

		When("the memory amount is invalid", func() {
			var qemuSystem qemu.QemuSystem

			BeforeEach(func() {
				qemuSystem = qemu.QemuSystem{
					Memory:   "3070M",
					CpuCount: 9,
					Numa:     3,
				}
			})

			It("fails", func() {
				_, err := qemuSystem.GenerateNumaArguments()
				Expect(err).To(MatchError("unable to calculate symmetric NUMA topology with vCPUs:9 Memory:3070M NUMA:3"))
			})
		})
	})
})
