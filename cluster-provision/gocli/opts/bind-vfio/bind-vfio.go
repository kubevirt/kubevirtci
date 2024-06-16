package bindvfio

import (
	"fmt"
	"strings"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type BindVfioOpt struct {
	sshPort uint16
	nodeIdx int
	pciID   string
}

func NewBindVfioOpt(sshPort uint16, nodeIdx int, id string) *BindVfioOpt {
	return &BindVfioOpt{
		sshPort: sshPort,
		nodeIdx: nodeIdx,
		pciID:   id,
	}
}

func (o *BindVfioOpt) Exec() error {
	addr, err := utils.JumpSSH(o.sshPort, o.nodeIdx, `lspci -D -d `+o.pciID, true, true)
	if err != nil {
		return err
	}

	pciDevId := strings.Split(addr, " ")[0]

	devSysfsPath := "/sys/bus/pci/devices/" + pciDevId
	driverPath := devSysfsPath + "/driver"
	driverOverride := devSysfsPath + "/driver_override"

	driver, err := utils.JumpSSH(o.sshPort, o.nodeIdx, "readlink "+driverPath+" | awk -F'/' '{print $NF}'", true, true)
	if err != nil {
		return err
	}

	cmds := []string{
		"if [[ ! -d /sys/bus/pci/devices/" + pciDevId + " ]]; then echo 'Error: PCI address does not exist!' && exit 1; fi",
		"if [[ ! -d /sys/bus/pci/devices/" + pciDevId + "/iommu/ ]]; then echo 'Error: No vIOMMU found in the VM' && exit 1; fi",
		"modprobe -i vfio-pci",
		"[[ " + driver + "!= 'vfio-pci' ]] && echo " + pciDevId + " > " + driverPath + "/unbind && echo 'vfio-pci' > " + driverOverride + " && echo " + pciDevId + " > /sys/bus/pci/drivers/vfio-pci/bind",
	}

	for _, cmd := range cmds {
		if _, err := utils.JumpSSH(o.sshPort, 1, cmd, true, true); err != nil {
			return err
		}
	}

	newDriver, err := utils.JumpSSH(o.sshPort, o.nodeIdx, "readlink "+driverPath+" | awk -F'/' '{print $NF}'", true, true)
	if err != nil {
		return err
	}
	if newDriver != "vfio-pci" {
		return fmt.Errorf("Error: Failed to bind to vfio-pci driver")
	}
	return nil
}
