package controlplane

import "kubevirt.io/kubevirtci/cluster-provision/gocli/cri"

type RunETCDPhase struct {
	dnsmasqID        string
	containerRuntime cri.ContainerClient
	pkiPath          string
}

func NewRunETCDPhase(dnsmasqID string, containerRuntime cri.ContainerClient, pkiPath string) Phase {
	return &RunETCDPhase{
		dnsmasqID:        dnsmasqID,
		containerRuntime: containerRuntime,
		pkiPath:          pkiPath,
	}
}

func (p *RunETCDPhase) Run() error {
	etcdImageRepo := registry + "/" + etcdImage
	err := p.containerRuntime.ImagePull(etcdImageRepo)
	if err != nil {
		return err
	}

	etcdCmd := []string{"etcd"}
	args := buildEtcdCmdArgs()

	for flag, value := range args {
		etcdCmd = append(etcdCmd, flag+"="+value)
	}

	createOpts := &cri.CreateOpts{
		Name: "etcd",
		Mounts: map[string]string{
			p.pkiPath: "/etc/kubernetes/pki/etcd",
		},
		Network: "container:" + p.dnsmasqID,
		Command: etcdCmd,
	}

	etcdContainer, err := p.containerRuntime.Create(etcdImageRepo, createOpts)
	if err != nil {
		return err
	}

	err = p.containerRuntime.Start(etcdContainer)
	if err != nil {
		return err
	}

	return nil
}
