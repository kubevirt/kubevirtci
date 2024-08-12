package bootc

import (
	"embed"
	"strings"

	"os"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
)

//go:embed k8s-container/k8s.Containerfile
var k8sContainerfile []byte

//go:embed k8s-container/linux.Containerfile
var linuxContainerfile []byte

//go:embed k8s-container/provision-system.sh
var provisionSystem []byte

//go:embed k8s-container/provision-system.service
var provisionSystemService []byte

//go:embed k8s-container/config.toml
var configToml []byte

//go:embed k8s-container/patches/*
var patches embed.FS

type BootcProvisioner struct {
	cri cri.ContainerClient
}

func NewBootcProvisioner(cri cri.ContainerClient) *BootcProvisioner {
	return &BootcProvisioner{
		cri: cri,
	}
}

func (b *BootcProvisioner) BuildLinuxBase(tag string) error {
	fileName := "provision-system.sh"
	containerFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = containerFile.Write(provisionSystem)
	if err != nil {
		return err
	}

	fileName = "provision-system.service"
	containerFile, err = os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = containerFile.Write(provisionSystemService)
	if err != nil {
		return err
	}

	fileName = "linux.Containerfile"
	containerFile, err = os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = containerFile.Write(linuxContainerfile)
	if err != nil {
		return err
	}

	err = b.cri.Build(tag, fileName, map[string]string{})
	if err != nil {
		return err
	}
	return nil
}

func (b *BootcProvisioner) BuildK8sBase(tag, k8sVersion, baseImage string) error {
	fileName := "k8s.Containerfile"
	fileWithBase := strings.Replace(string(k8sContainerfile), "LINUX_BASE", baseImage, 1)

	containerFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = containerFile.Write([]byte(fileWithBase))
	if err != nil {
		return err
	}
	_ = os.Mkdir("patches", 0777)

	err = b.cri.Build(tag, fileName, map[string]string{"VERSION": k8sVersion})
	if err != nil {
		return err
	}
	return nil
}

func (b *BootcProvisioner) GenerateQcow(image string) error {
	_ = os.Mkdir("output", 0777)

	configFileName := "config.toml"
	conf, err := os.Create(configFileName)
	if err != nil {
		return err
	}
	_, err = conf.Write(configToml)
	if err != nil {
		return err
	}

	runArgs := []string{"--rm",
		"--privileged",
		"--security-opt",
		"label=type:unconfined_t",
		"-v",
		"./output:/output",
		"-v",
		"/var/lib/containers/storage:/var/lib/containers/storage",
		"-v",
		"./config.toml:/config.toml:ro",
		"quay.io/centos-bootc/bootc-image-builder:latest",
		"--type",
		"qcow2",
		"--local",
		"localhost/" + image}

	err = b.cri.Run(runArgs)
	if err != nil {
		return err
	}

	return nil
}
