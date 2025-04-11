package swap

import (
	"fmt"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

type swapOpt struct {
	sshClient     libssh.Client
	swapiness     int
	unlimitedSwap bool
	size          int
}

func NewSwapOpt(sc libssh.Client, swapiness int, us bool, size int) *swapOpt {
	return &swapOpt{
		sshClient:     sc,
		swapiness:     swapiness,
		unlimitedSwap: us,
		size:          size,
	}
}

func (o *swapOpt) Exec() error {
	if o.size != 0 {
		if err := o.sshClient.Command("fallocate -l " + fmt.Sprintf("%dG", o.size) + " /swapfile"); err != nil {
			return err
		}
		if err := o.sshClient.Command("mkswap /swapfile"); err != nil {
			return err
		}
	}
	if err := o.sshClient.Command("swapon -a"); err != nil {
		return err
	}

	if o.swapiness != 0 {
		cmds := []string{
			"/bin/su -c \"echo vm.swappiness = " + fmt.Sprintf("%d", o.swapiness) + " >> /etc/sysctl.conf\"",
			"sysctl vm.swappiness=" + fmt.Sprintf("%d", o.swapiness),
		}
		for _, cmd := range cmds {
			if err := o.sshClient.Command(cmd); err != nil {
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
			if err := o.sshClient.Command(cmd); err != nil {
				return err
			}
		}
	}

	return nil
}
