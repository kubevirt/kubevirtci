package ksm

import (
	"fmt"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type KsmOpt struct {
	sshClient    utils.SSHClient
	sshPort      uint16
	scanInterval int
	pagesToScan  int
	nodeIdx      int
}

func NewKsmOpt(sc utils.SSHClient, sshPort uint16, idx, si, pages int) *KsmOpt {
	return &KsmOpt{
		sshClient:    sc,
		sshPort:      sshPort,
		scanInterval: si,
		pagesToScan:  pages,
		nodeIdx:      idx,
	}
}

func (o *KsmOpt) Exec() error {
	cmds := []string{
		"echo 1 | sudo tee /sys/kernel/mm/ksm/run >/dev/null",
		"echo " + fmt.Sprintf("%d", o.scanInterval) + " | sudo tee /sys/kernel/mm/ksm/sleep_millisecs >/dev/null",
		"echo " + fmt.Sprintf("%d", o.pagesToScan) + " | sudo tee /sys/kernel/mm/ksm/pages_to_scan >/dev/null",
	}

	for _, cmd := range cmds {
		if _, err := o.sshClient.JumpSSH(o.sshPort, 1, cmd, true, true); err != nil {
			return err
		}
	}

	return nil
}
