package registryproxy

import "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"

type RegistryProxyOpt struct {
	sshClient libssh.Client
	proxyUrl  string
}

func NewRegistryProxyOpt(sshClient libssh.Client, proxyUrl string) *RegistryProxyOpt {
	return &RegistryProxyOpt{
		sshClient: sshClient,
		proxyUrl:  proxyUrl,
	}
}

func (rp *RegistryProxyOpt) Exec() error {
	setupUrl := "http://" + rp.proxyUrl + ":3128/setup/systemd"
	cmds := []string{
		"curl " + setupUrl + " > proxyscript.sh",
		"sed s/docker.service/containerd.service/g proxyscript.sh",
		`sed '/Environment/ s/$/ \"NO_PROXY=127.0.0.0\/8,10.0.0.0\/8,172.16.0.0\/12,192.168.0.0\/16\"/ proxyscript.sh`,
		"source proxyscript.sh",
	}
	for _, cmd := range cmds {
		if _, err := rp.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}
	return nil

}
