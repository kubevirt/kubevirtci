package dockerproxy

import (
	_ "embed"
	"strings"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed conf/override.conf
var override []byte

type dockerProxyOpt struct {
	proxy     string
	sshClient libssh.Client
}

func NewDockerProxyOpt(sc libssh.Client, proxy string) *dockerProxyOpt {
	return &dockerProxyOpt{
		proxy:     proxy,
		sshClient: sc,
	}
}

func (o *dockerProxyOpt) Exec() error {
	script := strings.ReplaceAll(string(override), "$PROXY", o.proxy)

	cmds := []string{
		"curl " + o.proxy + "/ca.crt > /etc/pki/ca-trust/source/anchors/docker_registry_proxy.crt",
		"update-ca-trust",
		"mkdir -p /etc/systemd/system/crio.service.d",
		"echo '" + script + "' | sudo tee /etc/systemd/system/crio.service.d/override.conf > /dev/null",
		"systemctl daemon-reload",
		"systemctl restart crio.service",
	}

	for _, cmd := range cmds {
		if err := o.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}
