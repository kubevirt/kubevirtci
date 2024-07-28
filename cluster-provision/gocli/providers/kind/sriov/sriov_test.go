package sriov

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SR-IOV test suite")
}

var _ = Describe("SR-IOV functionality", func() {
	var (
		mockCtrl  *gomock.Controller
		sshClient *kubevirtcimocks.MockSSHClient
		ks        *KindSriov
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		sshClient = kubevirtcimocks.NewMockSSHClient(mockCtrl)
		ks = &KindSriov{}
		ks.vfsCount = 6
	})

	AfterEach(func() {
		mockCtrl.Finish()
		sshClient = nil
	})

	Describe("fetchNodePfs", func() {
		It("should execute the correct commands", func() {
			sshClient.EXPECT().Command("grep vfio_pci /proc/modules", false).Return("vfio-pci", nil)
			sshClient.EXPECT().Command("modprobe -i vfio_pci", true)
			sshClient.EXPECT().Command("find /sys/class/net/*/device/", false).Return("/sys/class/net/eth0/device", nil)

			_, err := ks.fetchNodePfs(sshClient)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("createVFsforPF", func() {
		It("should execute the correct commands", func() {
			pfSysFsPath := "/sys/devices/pci/pciaddr/net/eth0/device"
			sshClient.EXPECT().Command("readlink -e /sys/class/net/eth0/device", false).Return(pfSysFsPath, nil)
			sshClient.EXPECT().Command("cat /sys/devices/pci/pciaddr/net/eth0/device/sriov_totalvfs", false).Return("6", nil)
			sshClient.EXPECT().Command("echo 0 >> "+pfSysFsPath+"/sriov_numvfs", true)
			sshClient.EXPECT().Command("echo 6 >> "+pfSysFsPath+"/sriov_numvfs", true)
			sshClient.EXPECT().Command(`readlink -e `+pfSysFsPath+`/virtfn*`, false)

			_, err := ks.createVFsforPF(sshClient, "/sys/class/net/eth0/device")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("bindToVfio", func() {
		It("should execute the correct commands", func() {
			vfSysFsPath := "/sys/devices/pci/pciaddr/net/eth0/device/virtfn1"
			driverPath := vfSysFsPath + "/driver"
			driverOverride := vfSysFsPath + "/driver_override"

			sshClient.EXPECT().Command("basename "+vfSysFsPath, false).Return("virtfn1", nil)
			sshClient.EXPECT().Command("readlink "+driverPath+" | awk -F'/' '{print $NF}'", false).Return("ixgbevf", nil)
			sshClient.EXPECT().Command("basename ixgbevf", false).Return("ixgbevf", nil)
			sshClient.EXPECT().Command("[[ 'ixgbevf' != 'vfio-pci' ]] && echo virtfn1"+" > "+driverPath+"/unbind && echo 'vfio-pci' > "+driverOverride+" && echo virtfn1"+" > /sys/bus/pci/drivers/vfio-pci/bind", true)

			err := ks.bindToVfio(sshClient, vfSysFsPath)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
