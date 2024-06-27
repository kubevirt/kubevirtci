package swap

import (
	"fmt"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

type SwapOpt struct {
	sshClient     utils.SSHClient
	sshPort       uint16
	nodeIdx       int
	swapiness     int
	unlimitedSwap bool
	size          string
}

func NewSwapOpt(sc utils.SSHClient, sshPort uint16, idx int, swapiness int, us bool, size string) *SwapOpt {
	return &SwapOpt{
		sshClient:     sc,
		sshPort:       sshPort,
		nodeIdx:       idx,
		swapiness:     swapiness,
		unlimitedSwap: us,
		size:          size,
	}
}

func (o *SwapOpt) Exec() error {
	if o.size != "" {
		if _, err := o.sshClient.JumpSSH(o.sshPort, 1, "dd if=/dev/zero of=/swapfile count="+o.size+" bs=1G", true, true); err != nil {
			return err
		}
	}
	if _, err := o.sshClient.JumpSSH(o.sshPort, 1, "swapon -a", true, true); err != nil {
		return err
	}

	if o.swapiness != 0 {
		cmds := []string{
			"/bin/su -c \"echo vm.swappiness = " + fmt.Sprintf("%d", o.swapiness) + " >> /etc/sysctl.conf\"",
			"sysctl vm.swappiness=" + fmt.Sprintf("%d", o.swapiness),
		}
		for _, cmd := range cmds {
			if _, err := o.sshClient.JumpSSH(o.sshPort, 1, cmd, true, true); err != nil {
				return err
			}
		}
	}

	if o.unlimitedSwap {
		cmds := []string{
			`sed -i 's/memorySwap: {}/memorySwap:\n  swapBehavior: UnlimitedSwap/g' /var/lib/kubelet/config.yaml`,
			"systemctl restart kubelet",
		}
		for _, cmd := range cmds {
			if _, err := o.sshClient.JumpSSH(o.sshPort, 1, cmd, true, true); err != nil {
				return err
			}
		}
	}

	return nil
}
