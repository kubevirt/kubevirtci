package dockerproxy

import (
	"embed"
	"strings"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed conf/*
var f embed.FS

type DockerProxyOpt struct {
	proxy     string
	sshClient libssh.Client
}

func NewDockerProxyOpt(sc libssh.Client, proxy string) *DockerProxyOpt {
	return &DockerProxyOpt{
		proxy:     proxy,
		sshClient: sc,
	}
}

func (o *DockerProxyOpt) Exec() error {
	override, err := f.ReadFile("conf/override.conf")
	if err != nil {
		return err
	}
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
		if _, err := o.sshClient.Command(cmd, true); err != nil {
			return err
		}
	}

	return nil
}
