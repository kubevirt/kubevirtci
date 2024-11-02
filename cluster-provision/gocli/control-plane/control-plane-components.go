package controlplane

import "kubevirt.io/kubevirtci/cluster-provision/gocli/cri"

type RunControlPlaneComponentsPhase struct {
	dnsmasqID        string
	k8sVersion       string
	pkiPath          string
	containerRuntime cri.ContainerClient
}

func NewRunControlPlaneComponentsPhase(dnsmasqID string, containerRuntime cri.ContainerClient, pkiPath, k8sVersion string) *RunControlPlaneComponentsPhase {
	return &RunControlPlaneComponentsPhase{
		dnsmasqID:        dnsmasqID,
		pkiPath:          pkiPath,
		containerRuntime: containerRuntime,
		k8sVersion:       k8sVersion,
	}
}

func (p *RunControlPlaneComponentsPhase) Run() error {
	componentFuncs := []func() error{p.runApiServer, p.runControllerMgr, p.runScheduler}
	for _, componentFunc := range componentFuncs {
		err := componentFunc()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *RunControlPlaneComponentsPhase) runApiServer() error {
	apiServerImage := registry + "/" + apiServer + ":" + p.k8sVersion
	err := p.containerRuntime.ImagePull(apiServerImage)
	if err != nil {
		return err
	}

	args := buildApiServerCmdArgs()

	cmd := []string{"kube-apiserver"}
	for flag, values := range args {
		cmd = append(cmd, flag+"="+values)
	}

	createOpts := &cri.CreateOpts{
		Name: "api-server",
		Mounts: map[string]string{
			p.pkiPath: "/etc/kubernetes/pki/",
		},
		Network: "container:" + p.dnsmasqID,
		Command: cmd,
	}

	apiserverContainer, err := p.containerRuntime.Create(apiServerImage, createOpts)
	if err != nil {
		return err
	}

	err = p.containerRuntime.Start(apiserverContainer)
	if err != nil {
		return err
	}
	return nil
}

func (p *RunControlPlaneComponentsPhase) runControllerMgr() error {
	ctrlMgrImage := registry + "/" + controllerManager + ":" + p.k8sVersion
	err := p.containerRuntime.ImagePull(ctrlMgrImage)
	if err != nil {
		return err
	}

	args := buildControllerMgrCmdArgs()

	cmd := []string{"kube-controller-manager"}
	for flag, values := range args {
		cmd = append(cmd, flag+"="+values)
	}

	createOpts := &cri.CreateOpts{
		Name: "kube-controller-manager",
		Mounts: map[string]string{
			p.pkiPath: "/etc/kubernetes/pki/",
		},
		Network: "container:" + p.dnsmasqID,
		Command: cmd,
	}

	apiserverContainer, err := p.containerRuntime.Create(ctrlMgrImage, createOpts)
	if err != nil {
		return err
	}

	err = p.containerRuntime.Start(apiserverContainer)
	if err != nil {
		return err
	}
	return nil
}

func (p *RunControlPlaneComponentsPhase) runScheduler() error {
	schedulerImage := registry + "/" + scheduler + ":" + p.k8sVersion
	err := p.containerRuntime.ImagePull(schedulerImage)
	if err != nil {
		return err
	}

	cmd := []string{"kube-scheduler", "--kubeconfig=/etc/kubernetes/kube-scheduler.kubeconfig"}

	createOpts := &cri.CreateOpts{
		Name: "kube-scheduler",
		Mounts: map[string]string{
			p.pkiPath: "/etc/kubernetes/pki/",
		},
		Network: "container:" + p.dnsmasqID,
		Command: cmd,
	}

	apiserverContainer, err := p.containerRuntime.Create(schedulerImage, createOpts)
	if err != nil {
		return err
	}

	err = p.containerRuntime.Start(apiserverContainer)
	if err != nil {
		return err
	}
	return nil
}
