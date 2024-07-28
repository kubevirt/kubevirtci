package network

import "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"

type NetworkOpt struct {
	sshClient libssh.Client
}

func NewNetworkOpt(sshClient libssh.Client) *NetworkOpt {
	return &NetworkOpt{
		sshClient: sshClient,
	}
}

func (n *NetworkOpt) Exec() error {
	cmds := []string{
		"modprobe br_netfilter",
		"sysctl -w net.bridge.bridge-nf-call-arptables=1",
		"sysctl -w net.bridge.bridge-nf-call-iptables=1",
		"sysctl -w net.bridge.bridge-nf-call-ip6tables=1",
	}

	for _, cmd := range cmds {
		if _, err := n.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}
	return nil
}
