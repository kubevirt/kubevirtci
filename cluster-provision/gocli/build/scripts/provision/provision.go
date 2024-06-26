package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	kubeVirtCIDir := "/var/lib/kubevirtci"

	err := os.Mkdir(kubeVirtCIDir, 0755)
	if err != nil {
		panic(err)
	}

	err = os.Setenv("ISTIO_VERSION", "1.15.0")
	if err != nil {
		panic(err)
	}

	filePath := kubeVirtCIDir + "/shared_vars.sh"

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// why do we need to export those vars twice ?
	content := `#!/bin/bash
set -ex
export KUBELET_CGROUP_ARGS="--cgroup-driver=systemd --runtime-cgroups=/systemd/system.slice --kubelet-cgroups=/systemd/system.slice"
export ISTIO_BIN_DIR="/opt/istio-${ISTIO_VERSION}/bin"
`
	_, err = file.WriteString(content)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("/bin/bash", "-c", filePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	kernelVersion, err := runCMD("uname -r")
	if err != nil {
		panic(err)
	}

	_, err = runCMD("if growpart /dev/vda 1; then resize2fs /dev/vda1 fi")
	if err != nil {
		panic(err)
	}

	packages := []string{fmt.Sprintf("kernel-modules-%s", kernelVersion), "cloud-utils-growpart", "patch",
		"iscsi-initiator-utils", "nftables", "lvm2",
		"iproute-tc", "container-selinux", "libseccomp-devel",
		"centos-release-nfv-openvswitch", "openvswitch2.16", "NetworkManager",
		"NetworkManager-ovs", "NetworkManager-config-server"}

	for _, p := range packages {
		_, err = runCMD(fmt.Sprintf("dnf install -y %s", p))
		if err != nil {
			panic(err)
		}
	}

	f, err := os.OpenFile("/etc/udev/rules.d/60-force-ssd-rotational.rules", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(`ACTION=="add|change", SUBSYSTEM=="block", KERNEL=="vd[a-z]", ATTR{queue/rotational}="0"`); err != nil {
		panic(err)
	}

	istioBinDir := os.Getenv("ISTIO_BIN_DIR")
	if istioBinDir == "" {
		istioBinDir = "/opt/istio"
	}

	err = os.MkdirAll(istioBinDir, 0755)
	if err != nil {
		panic(err)
	}

	istioctlURL := fmt.Sprintf("https://storage.googleapis.com/kubevirtci-istioctl-mirror/istio-%s/bin/istioctl", os.Getenv("ISTIO_VERSION"))
	response, err := http.Get(istioctlURL)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	out, err := os.Create(filepath.Join(istioBinDir, "istioctl"))
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		panic(err)
	}

	// Make the istioctl binary executable
	err = os.Chmod(filepath.Join(istioBinDir, "istioctl"), 0755)
	if err != nil {
		panic(err)
	}
	log.Println("Executed linux phase successfully")
}

func runCMD(cmd string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	command := exec.Command(cmd)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	if err != nil {
		return "", fmt.Errorf(stderr.String())
	}
	return stdout.String(), nil
}
