package dockerproxy

import (
	"embed"
	"strings"

	utils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

//go:embed conf/*
var f embed.FS

type DockerProxyOpt struct {
	sshPort uint16
	proxy   string
	nodeIdx int
}

func NewDockerProxyOpt(port uint16, proxy string, idx int) *DockerProxyOpt {
	return &DockerProxyOpt{
		sshPort: port,
		proxy:   proxy,
		nodeIdx: idx,
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
		if _, err := utils.JumpSSH(o.sshPort, o.nodeIdx, cmd, true, true); err != nil {
			return err
		}
	}

	return nil
}
