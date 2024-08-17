package k8sprovision

import (
	"embed"
	"encoding/base64"
	"fmt"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

//go:embed conf/*
var f embed.FS

type k8sProvisioner struct {
	version   string
	slim      bool
	sshClient libssh.Client
}

func NewK8sProvisioner(sshClient libssh.Client, version string, slim bool) *k8sProvisioner {
	return &k8sProvisioner{
		version:   version,
		slim:      slim,
		sshClient: sshClient,
	}
}

func (k *k8sProvisioner) Exec() error {
	k8sProvision, err := f.ReadFile("conf/k8s_provision.sh")
	if err != nil {
		return err
	}

	kubeadm, err := f.ReadFile("conf/kubeadm.conf")
	if err != nil {
		return err
	}

	kubeadmIpv6, err := f.ReadFile("conf/kubeadm_ipv6.conf")
	if err != nil {
		return err
	}

	fetchImages, err := f.ReadFile("conf/fetch-images.sh")
	if err != nil {
		return err
	}

	extraImages, err := f.ReadFile("conf/extra-images")
	if err != nil {
		return err
	}

	cniDiff, err := f.ReadFile("conf/cni.diff")
	if err != nil {
		return err
	}

	cni6Diff, err := f.ReadFile("conf/cni_ipv6.diff")
	if err != nil {
		return err
	}

	cmds := []string{
		"echo " + base64.StdEncoding.EncodeToString(kubeadm) + " | base64 -d | tee /tmp/kubeadm.conf",
		"echo " + base64.StdEncoding.EncodeToString(kubeadmIpv6) + " | base64 -d | tee /tmp/kubeadm_ipv6.conf",
		"echo " + base64.StdEncoding.EncodeToString(fetchImages) + " | base64 -d | tee /tmp/fetch-images.sh",
		"echo " + base64.StdEncoding.EncodeToString(extraImages) + " | base64 -d | tee /tmp/extra-pre-pull-images",
		"echo " + base64.StdEncoding.EncodeToString(cniDiff) + " | base64 -d | tee /tmp/cni.diff",
		"echo " + base64.StdEncoding.EncodeToString(cni6Diff) + " | base64 -d | tee /tmp/cni_ipv6.diff",
		"echo " + base64.StdEncoding.EncodeToString(k8sProvision) + " | base64 -d | tee /tmp/k8s_provision.sh",
		fmt.Sprintf("sudo version=%s slim=%t bash /tmp/k8s_provision.sh", k.version, k.slim),
	}

	for _, cmd := range cmds {
		if err := k.sshClient.Command(cmd); err != nil {
			return err
		}
	}

	return nil
}
