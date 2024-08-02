package bindvfio

import (
	"github.com/sirupsen/logrus"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func AddExpectCalls(sshClient *kubevirtcimocks.MockSSHClient, pciID string) {
	sshClient.EXPECT().CommandWithNoStdOut("lspci -D -d "+pciID).Return("testpciaddr something something", nil)

	devSysfsPath := "/sys/bus/pci/devices/testpciaddr"
	driverPath := devSysfsPath + "/driver"
	driverOverride := devSysfsPath + "/driver_override"

	sshClient.EXPECT().CommandWithNoStdOut("readlink "+driverPath+" | awk -F'/' '{print $NF}'").Return("not-vfio", nil)
	sshClient.EXPECT().Command("modprobe -i vfio-pci")
	sshClient.EXPECT().Command("ls /sys/bus/pci/drivers/vfio-pci")

	cmds := []string{
		"if [[ ! -d /sys/bus/pci/devices/testpciaddr ]]; then echo 'Error: PCI address does not exist!' && exit 1; fi",
		"if [[ ! -d /sys/bus/pci/devices/testpciaddr/iommu/ ]]; then echo 'Error: No vIOMMU found in the VM' && exit 1; fi",
		"[[ 'not-vfio' != 'vfio-pci' ]] && echo testpciaddr > " + driverPath + "/unbind && echo 'vfio-pci' > " + driverOverride + " && echo testpciaddr > /sys/bus/pci/drivers/vfio-pci/bind",
	}
	for _, cmd := range cmds {
		sshClient.EXPECT().Command(cmd)
	}
	logrus.Info("Added expect calls for soundcard ", pciID)
}
