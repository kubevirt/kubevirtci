package bindvfio

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	kubevirtcimocks "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/mock"
)

func TestBindVfio(t *testing.T) {
	sshClient := kubevirtcimocks.NewMockSSHClient(gomock.NewController(t))
	opt := NewBindVfioOpt(sshClient, 2020, 1, "8086:2668")

	sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, "lspci -D -d "+opt.pciID, true, false).Return("testpciaddr something something", nil)

	devSysfsPath := "/sys/bus/pci/devices/testpciaddr"
	driverPath := devSysfsPath + "/driver"
	driverOverride := devSysfsPath + "/driver_override"

	sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, "readlink "+driverPath+" | awk -F'/' '{print $NF}'", true, false).Return("not-vfio", nil)
	sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, "modprobe -i vfio-pci", true, false)
	sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, "ls /sys/bus/pci/drivers/vfio-pci", true, false)

	cmds := []string{
		"if [[ ! -d /sys/bus/pci/devices/testpciaddr ]]; then echo 'Error: PCI address does not exist!' && exit 1; fi",
		"if [[ ! -d /sys/bus/pci/devices/testpciaddr/iommu/ ]]; then echo 'Error: No vIOMMU found in the VM' && exit 1; fi",
		"[[ 'not-vfio' != 'vfio-pci' ]] && echo testpciaddr > " + driverPath + "/unbind && echo 'vfio-pci' > " + driverOverride + " && echo testpciaddr > /sys/bus/pci/drivers/vfio-pci/bind",
	}
	for _, cmd := range cmds {
		sshClient.EXPECT().JumpSSH(opt.sshPort, opt.nodeIdx, cmd, true, true)
	}

	err := opt.Exec()
	assert.NoError(t, err)
}
