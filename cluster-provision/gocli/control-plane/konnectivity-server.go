package controlplane

import (
	"fmt"
	"os"
	"path"

	"k8s.io/client-go/tools/clientcmd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
)

type KonnectivityPhase struct {
	pkiPath          string
	dnsmasqID        string
	serverImage      string
	k8sVersion       string
	containerRuntime cri.ContainerClient
}

func NewKonnectivityPhase(cr cri.ContainerClient, pkiPath, serverImage, dnsmasqID, k8sVersion string) *KonnectivityPhase {
	return &KonnectivityPhase{
		dnsmasqID:        dnsmasqID,
		pkiPath:          pkiPath,
		containerRuntime: cr,
		serverImage:      serverImage,
		k8sVersion:       k8sVersion,
	}
}

func (k *KonnectivityPhase) Run() error {
	if err := k.createKonnectivityKubeConfig(); err != nil {
		return err
	}

	img := registry + "/" + k.serverImage
	err := k.containerRuntime.ImagePull(img)
	if err != nil {
		return err
	}

	args := buildKonnectivityArgs()

	cmd := []string{"konnectivity-server"}
	for flag, values := range args {
		cmd = append(cmd, flag+"="+values)
	}

	createOpts := &cri.CreateOpts{
		Name: "k8s-" + k.k8sVersion + "-konnectivity",
		Mounts: map[string]string{
			k.pkiPath: "/etc/kubernetes/pki/",
		},
		Network: "container:" + k.dnsmasqID,
		Command: cmd,
	}

	konnectivityContainer, err := k.containerRuntime.Create(img, createOpts)
	if err != nil {
		return err
	}

	err = k.containerRuntime.Start(konnectivityContainer)
	if err != nil {
		return err
	}
	return nil
}

func (k *KonnectivityPhase) createKonnectivityKubeConfig() error {
	ca, err := os.ReadFile(path.Join(k.pkiPath, "ca.crt"))
	if err != nil {
		return err
	}

	clientCert, err := os.ReadFile(path.Join(k.pkiPath, "konnectivity.crt"))
	if err != nil {
		return err
	}

	key, err := os.ReadFile(path.Join(k.pkiPath, "konnectivity.pem"))
	if err != nil {
		return err
	}

	kubeconfig := buildKubeConfigFromCerts(ca, clientCert, key, "https://127.0.0.1:6443", "system:konnectivity-server")
	err = clientcmd.WriteToFile(kubeconfig, k.pkiPath+"/konnectivity/.kubeconfig")
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}
