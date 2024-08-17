package providers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alessio/shellescape"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/k8sprovision"
	provisionopt "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/provision"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/rootkey"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

func (kp *KubevirtProvider) Provision(ctx context.Context, cancel context.CancelFunc, portMap nat.PortMap, k8sVersion string) (retErr error) {
	prefix := fmt.Sprintf("k8s-%s-provision", kp.Version)
	target := fmt.Sprintf("quay.io/kubevirtci/k8s-%s", kp.Version)
	if kp.Phases == "linux" {
		target = kp.Image + "-base"
	}
	kp.Version = prefix

	stop := make(chan error, 10)
	containers, volumes, done := docker.NewCleanupHandler(kp.Docker, stop, os.Stdout, true)

	defer func() {
		stop <- retErr
		<-done
	}()

	go kp.handleInterrupt(cancel, stop)

	err := docker.ImagePull(kp.Docker, ctx, kp.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	dnsmasq, err := kp.runDNSMasq(ctx, portMap)
	if err != nil {
		return err
	}

	kp.DNSMasq = dnsmasq
	containers <- dnsmasq

	dm, err := kp.Docker.ContainerInspect(context.Background(), dnsmasq)
	if err != nil {
		return err
	}

	sshPort, err := utils.GetPublicPort(utils.PortSSH, dm.NetworkSettings.Ports)
	if err != nil {
		return err
	}

	nodeName := kp.nodeNameFromIndex(1)
	nodeNum := fmt.Sprintf("%02d", 1)

	vol, err := kp.Docker.VolumeCreate(ctx, volume.CreateOptions{
		Name: fmt.Sprintf("%s-%s", prefix, nodeName),
	})
	if err != nil {
		return err
	}
	volumes <- vol.Name
	registryVol, err := kp.Docker.VolumeCreate(ctx, volume.CreateOptions{
		Name: fmt.Sprintf("%s-%s", prefix, "registry"),
	})
	if err != nil {
		return err
	}

	node, err := kp.Docker.ContainerCreate(ctx, &container.Config{
		Image: kp.Image,
		Env: []string{
			fmt.Sprintf("NODE_NUM=%s", nodeNum),
		},
		Volumes: map[string]struct{}{
			"/var/run/disk":     {},
			"/var/lib/registry": {},
		},
		Cmd: []string{"/bin/bash", "-c", fmt.Sprintf("/vm.sh --memory %s --cpu %s %s", kp.Memory, strconv.Itoa(int(kp.CPU)), kp.QemuArgs)},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "volume",
				Source: vol.Name,
				Target: "/var/run/disk",
			},
			{
				Type:   "volume",
				Source: registryVol.Name,
				Target: "/var/lib/registry",
			},
		},
		Privileged:  true,
		NetworkMode: container.NetworkMode("container:" + kp.DNSMasq),
	}, nil, nil, kp.nodeContainer(kp.Version, nodeName))
	if err != nil {
		return err
	}
	containers <- node.ID
	if err := kp.Docker.ContainerStart(ctx, node.ID, container.StartOptions{}); err != nil {
		return err
	}

	// Wait for ssh.sh script to exist
	_, err = docker.Exec(kp.Docker, kp.nodeContainer(kp.Version, nodeName), []string{"bin/bash", "-c", "while [ ! -f /ssh_ready ] ; do sleep 1; done", "checking for ssh.sh script"}, os.Stdout)
	if err != nil {
		return err
	}

	sshClient, err := libssh.NewSSHClient(sshPort, 1, false)
	if err != nil {
		return err
	}

	rootkey := rootkey.NewRootKey(sshClient)
	if err = rootkey.Exec(); err != nil {
		fmt.Println(err)
	}

	sshClient, err = libssh.NewSSHClient(sshPort, 1, true)
	if err != nil {
		return err
	}

	if strings.Contains(kp.Phases, "linux") {
		provisionOpt := provisionopt.NewLinuxProvisioner(sshClient)
		if err = provisionOpt.Exec(); err != nil {
			return err
		}
	}

	if strings.Contains(kp.Phases, "k8s") {
		// copy provider scripts
		if err = sshClient.Command("mkdir -p /tmp/ceph /tmp/cnao /tmp/nfs-csi /tmp/nodeports /tmp/prometheus /tmp/whereabouts /tmp/kwok"); err != nil {
			return err
		}
		// Copy manifests to the VM
		success, err := docker.Exec(kp.Docker, kp.nodeContainer(kp.Version, nodeName), []string{"/bin/bash", "-c", "scp -r -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i vagrant.key -P 22 /scripts/manifests/* root@192.168.66.101:/tmp"}, os.Stdout)
		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("error copying manifests to node")
		}

		provisionK8sOpt := k8sprovision.NewK8sProvisioner(sshClient, k8sVersion, kp.Slim)
		if err = provisionK8sOpt.Exec(); err != nil {
			return err
		}
	}

	_ = sshClient.Command("sudo shutdown now -h")

	_, err = docker.Exec(kp.Docker, kp.nodeContainer(kp.Version, nodeName), []string{"rm", "/ssh_ready"}, io.Discard)
	if err != nil {
		return err
	}

	logrus.Info("waiting for the node to stop")
	okChan, errChan := kp.Docker.ContainerWait(ctx, kp.nodeContainer(kp.Version, nodeName), container.WaitConditionNotRunning)
	select {
	case <-okChan:
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("waiting for the node to stop failed: %v", err)
		}
	}

	if len(kp.AdditionalKernelArgs) > 0 {
		dir, err := os.MkdirTemp("", "gocli")
		if err != nil {
			return fmt.Errorf("failed creating a temporary directory: %v", err)
		}
		defer os.RemoveAll(dir)
		if err := os.WriteFile(filepath.Join(dir, "additional.kernel.args"), []byte(shellescape.QuoteCommand(kp.AdditionalKernelArgs)), 0666); err != nil {
			return fmt.Errorf("failed creating additional.kernel.args file: %v", err)
		}
		if err := kp.copyDirectory(ctx, kp.Docker, node.ID, dir, "/"); err != nil {
			return fmt.Errorf("failed copying additional kernel arguments into the container: %v", err)
		}
	}

	logrus.Infof("Commiting the node as %s", target)
	_, err = kp.Docker.ContainerCommit(ctx, node.ID, container.CommitOptions{
		Reference: target,
		Comment:   "PROVISION SUCCEEDED",
		Author:    "gocli",
		Changes:   nil,
		Pause:     false,
		Config:    nil,
	})
	if err != nil {
		return fmt.Errorf("commiting the node failed: %v", err)
	}

	return nil
}
