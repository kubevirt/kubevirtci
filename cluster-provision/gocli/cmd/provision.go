package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	containers2 "kubevirt.io/kubevirtci/cluster-provision/gocli/containers"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

// NewProvisionCommand provision given cluster
func NewProvisionCommand() *cobra.Command {

	provision := &cobra.Command{
		Use:   "provision",
		Short: "provision starts a given cluster",
		RunE:  provisionCluster,
		Args:  cobra.ExactArgs(1),
	}
	provision.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	provision.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	provision.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	provision.Flags().Bool("random-ports", false, "expose all ports on random localhost ports")
	provision.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	provision.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")
	provision.Flags().Bool("cgroupv2", false, "set UNIFIED_CGROUP_HIERARCHY environment variable for the provision script")
	provision.Flags().String("container-suffix", "", "use additional suffix for the provisioned container image")

	return provision
}

func provisionCluster(cmd *cobra.Command, args []string) (retErr error) {
	packagePath := args[0]
	versionBytes, err := ioutil.ReadFile(filepath.Join(packagePath, "version"))
	if err != nil {
		return err
	}
	version := strings.TrimSpace(string(versionBytes))
	baseBytes, err := ioutil.ReadFile(filepath.Join(packagePath, "base"))
	if err != nil {
		return err
	}
	base := fmt.Sprintf("quay.io/kubevirtci/%s", strings.TrimSpace(string(baseBytes)))

	containerSuffix, err := cmd.Flags().GetString("container-suffix")
	if err != nil {
		return err
	}
	name := filepath.Base(packagePath)
	if len(containerSuffix) > 0 {
		name = fmt.Sprintf("%s-%s", name, containerSuffix)
	}
	prefix := fmt.Sprintf("k8s-%s-provision", name)
	target := fmt.Sprintf("quay.io/kubevirtci/k8s-%s", name)
	scripts := filepath.Join(packagePath)

	memory, err := cmd.Flags().GetString("memory")
	if err != nil {
		return err
	}

	randomPorts, err := cmd.Flags().GetBool("random-ports")
	if err != nil {
		return err
	}

	portMap := nat.PortMap{}

	utils.AppendIfExplicit(portMap, utils.PortSSH, cmd.Flags(), "ssh-port")
	utils.AppendIfExplicit(portMap, utils.PortVNC, cmd.Flags(), "vnc-port")

	qemuArgs, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}

	cpu, err := cmd.Flags().GetUint("cpu")
	if err != nil {
		return err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	stop := make(chan error, 10)
	containers, volumes, done := docker.NewCleanupHandler(cli, stop, cmd.OutOrStderr(), true)

	defer func() {
		stop <- retErr
		<-done
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
		stop <- fmt.Errorf("Interrupt received, clean up")
	}()

	// Pull the base image
	err = docker.ImagePull(cli, ctx, base, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	// Start dnsmasq
	dnsmasq, err := containers2.DNSMasq(cli, ctx, &containers2.DNSMasqOptions{
		ClusterImage:       base,
		SecondaryNicsCount: 0,
		RandomPorts:        randomPorts,
		PortMap:            portMap,
		Prefix:             prefix,
		NodeCount:          1,
	})
	if err != nil {
		return err
	}
	containers <- dnsmasq.ID
	if err := cli.ContainerStart(ctx, dnsmasq.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	nodeName := nodeNameFromIndex(1)
	nodeNum := fmt.Sprintf("%02d", 1)

	vol, err := cli.VolumeCreate(ctx, volume.VolumeCreateBody{
		Name: fmt.Sprintf("%s-%s", prefix, nodeName),
	})
	if err != nil {
		return err
	}
	volumes <- vol.Name
	if len(qemuArgs) > 0 {
		qemuArgs = "--qemu-args " + qemuArgs
	}
	node, err := cli.ContainerCreate(ctx, &container.Config{
		Image: base,
		Env: []string{
			fmt.Sprintf("NODE_NUM=%s", nodeNum),
		},
		Volumes: map[string]struct{}{
			"/var/run/disk/": {},
		},
		Cmd: []string{"/bin/bash", "-c", fmt.Sprintf("/vm.sh --memory %s --cpu %s %s", memory, strconv.Itoa(int(cpu)), qemuArgs)},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "volume",
				Source: vol.Name,
				Target: "/var/run/disk",
			},
		},
		Privileged:  true,
		NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
	}, nil, nil, nodeContainer(prefix, nodeName))
	if err != nil {
		return err
	}
	containers <- node.ID
	if err := cli.ContainerStart(ctx, node.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Copy scripts
	srcInfo, err := archive.CopyInfoSourcePath(scripts, false)
	if err != nil {
		return err
	}

	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return err
	}
	defer srcArchive.Close()

	dstInfo := archive.CopyInfo{Path: "/scripts"}

	dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
	if err != nil {
		return err
	}
	defer preparedArchive.Close()

	err = cli.CopyToContainer(ctx, node.ID, dstDir, preparedArchive, types.CopyToContainerOptions{AllowOverwriteDirWithFile: false})
	if err != nil {
		return err
	}

	// Wait for ssh.sh script to exist
	err = _cmd(cli, nodeContainer(prefix, nodeName), "while [ ! -f /ssh_ready ] ; do sleep 1; done", "checking for ssh.sh script")
	if err != nil {
		return err
	}

	// Wait for the VM to be up
	err = _cmd(cli, nodeContainer(prefix, nodeName), "ssh.sh echo VM is up", "waiting for node to come up")
	if err != nil {
		return err
	}

	// Copy scripts to the VM
	err = _cmd(cli, nodeContainer(prefix, nodeName), "scp -r -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i vagrant.key -P 22 /scripts/manifests/* vagrant@192.168.66.101:/tmp", "copying scripts to the VM")
	if err != nil {
		return err
	}

	//check if we have a special provision script
	err = _cmd(cli, nodeContainer(prefix, nodeName), "test -f /scripts/provision.sh", "checking for provision script")
	if err != nil {
		return err
	}

	cgroupv2, err := cmd.Flags().GetBool("cgroupv2")
	if err != nil {
		return err
	}

	envVars := fmt.Sprintf("version=%s", version)
	if cgroupv2 {
		envVars = fmt.Sprintf("%s UNIFIED_CGROUP_HIERARCHY=1", envVars)
	}

	err = _cmd(cli, nodeContainer(prefix, nodeName), fmt.Sprintf("ssh.sh sudo %s /bin/bash < /scripts/provision.sh", envVars), "provisioning the node")
	if err != nil {
		return err
	}

	err = _cmd(cli, nodeContainer(prefix, nodeName), "ssh.sh sudo shutdown -h", "shutting down the node")
	if err != nil {
		return err
	}
	err = _cmd(cli, nodeContainer(prefix, nodeName), "rm /usr/local/bin/ssh.sh", "removing the ssh.sh script")
	if err != nil {
		return err
	}
	err = _cmd(cli, nodeContainer(prefix, nodeName), "rm /ssh_ready", "removing the ssh_ready mark")
	if err != nil {
		return err
	}
	logrus.Info("waiting for the node to stop")
	okChan, errChan := cli.ContainerWait(ctx, nodeContainer(prefix, nodeName), container.WaitConditionNotRunning)
	select {
	case <-okChan:
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("waiting for the node to stop failed: %v", err)
		}
	}

	logrus.Infof("Commiting the node as %s", target)
	_, err = cli.ContainerCommit(ctx, node.ID, types.ContainerCommitOptions{
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

func _cmd(cli *client.Client, container string, cmd string, description string) error {
	logrus.Info(description)
	success, err := docker.Exec(cli, container, []string{"/bin/bash", "-c", cmd}, os.Stdout)
	if err != nil {
		return fmt.Errorf("%s failed: %v", description, err)
	} else if !success {
		return fmt.Errorf("%s failed", cmd)
	}
	return nil
}
