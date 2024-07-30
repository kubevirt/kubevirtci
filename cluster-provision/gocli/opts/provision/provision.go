package provision

import (
	"embed"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed conf/*
var f embed.FS

type LinuxProvisioner struct {
	sshPort   uint16
	sshClient libssh.Client
}

func NewLinuxProvisioner(sc libssh.Client) *LinuxProvisioner {
	return &LinuxProvisioner{
		sshClient: sc,
	}
}

func (l *LinuxProvisioner) Exec() error {
	sharedVars, err := f.ReadFile("conf/shared_vars.sh")
	if err != nil {
		return nil
	}

	cmds := []string{
		`mkdir -p /var/lib/kubevirtci && echo '` + string(sharedVars) + `' |  tee /var/lib/kubevirtci/shared_vars.sh > /dev/null`,
		`dnf install -y "kernel-modules-$(uname -r)"`,
		"dnf install -y cloud-utils-growpart",
		`if growpart /dev/vda 1; then  resize2fs /dev/vda1; fi`,
		"dnf install -y patch",
		"systemctl stop firewalld || :",
		"systemctl disable firewalld || :",
		"dnf -y remove firewalld",
		"dnf -y install iscsi-initiator-utils",
		"dnf -y install nftables",
		"dnf -y install lvm2",
		`echo 'ACTION=="add|change", SUBSYSTEM=="block", KERNEL=="vd[a-z]", ATTR{queue/rotational}="0"' > /etc/udev/rules.d/60-force-ssd-rotational.rules`,
		"dnf install -y iproute-tc",
		"mkdir -p /opt/istio-1.15.0/bin",
		`curl "https://storage.googleapis.com/kubevirtci-istioctl-mirror/istio-1.15.0/bin/istioctl" -o "/opt/istio-1.15.0/bin/istioctl"`,
		`chmod +x /opt/istio-1.15.0/bin/istioctl`,
		"dnf install -y container-selinux",
		"dnf install -y libseccomp-devel",
		"dnf install -y centos-release-nfv-openvswitch",
		"dnf install -y openvswitch2.16",
		"dnf install -y NetworkManager NetworkManager-ovs NetworkManager-config-server",
	}
	for _, cmd := range cmds {
		if _, err := l.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}
	return nil
}
