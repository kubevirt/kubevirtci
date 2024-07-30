package ksm

import (
	"fmt"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type KsmOpt struct {
	sshClient    libssh.Client
	scanInterval int
	pagesToScan  int
}

func NewKsmOpt(sc libssh.Client, si, pages int) *KsmOpt {
	return &KsmOpt{
		sshClient:    sc,
		scanInterval: si,
		pagesToScan:  pages,
	}
}

func (o *KsmOpt) Exec() error {
	cmds := []string{
		"echo 1 | sudo tee /sys/kernel/mm/ksm/run >/dev/null",
		"echo " + fmt.Sprintf("%d", o.scanInterval) + " | sudo tee /sys/kernel/mm/ksm/sleep_millisecs >/dev/null",
		"echo " + fmt.Sprintf("%d", o.pagesToScan) + " | sudo tee /sys/kernel/mm/ksm/pages_to_scan >/dev/null",
	}

	for _, cmd := range cmds {
		if _, err := o.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}

	return nil
}
