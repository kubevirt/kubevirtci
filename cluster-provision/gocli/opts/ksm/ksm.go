package ksm

import (
	"fmt"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type ksmOpt struct {
	sshClient    libssh.Client
	scanInterval int
	pagesToScan  int
}

func NewKsmOpt(sc libssh.Client, si, pages int) *ksmOpt {
	return &ksmOpt{
		sshClient:    sc,
		scanInterval: si,
		pagesToScan:  pages,
	}
}

func (o *ksmOpt) Exec() error {
	cmds := []string{
		"echo 1 | sudo tee /sys/kernel/mm/ksm/run >/dev/null",
		"echo " + fmt.Sprintf("%d", o.scanInterval) + " | sudo tee /sys/kernel/mm/ksm/sleep_millisecs >/dev/null",
		"echo " + fmt.Sprintf("%d", o.pagesToScan) + " | sudo tee /sys/kernel/mm/ksm/pages_to_scan >/dev/null",
	}

	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}
