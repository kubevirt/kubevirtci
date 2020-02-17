package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

// NewProvisionCommand provision given cluster
func NewProvisionCommand() *cobra.Command {

	provision := &cobra.Command{
		Use:   "provision",
		Short: "provision starts a given cluster",
		RunE:  provision,
		Args:  cobra.ExactArgs(1),
	}
	provision.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	provision.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	provision.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	provision.Flags().String("scripts", "", "location for the provision and run scripts")
	provision.Flags().String("tag", "", "tag name for the commited image")
	provision.Flags().Bool("random-ports", false, "expose all ports on random localhost ports")
	provision.Flags().Bool("crio", false, "Use CRIO")
	provision.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	provision.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")
	provision.Flags().String("k8s-version", "", "kubernetes version to provision")

	return provision
}

func provision(cmd *cobra.Command, args []string) error {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return fmt.Errorf("failed getting prefix: %v", err)
	}

	scripts, err := cmd.Flags().GetString("scripts")
	if err != nil {
		return fmt.Errorf("failed getting scripts: %v", err)
	}

	tag, err := cmd.Flags().GetString("tag")
	if err != nil {
		return fmt.Errorf("failed getting tag: %v", err)
	}

	memory, err := cmd.Flags().GetString("memory")
	if err != nil {
		return fmt.Errorf("failed getting memory: %v", err)
	}

	randomPorts, err := cmd.Flags().GetBool("random-ports")
	if err != nil {
		return fmt.Errorf("failed getting random-ports: %v", err)
	}

	crio, err := cmd.Flags().GetBool("crio")
	if err != nil {
		return fmt.Errorf("failed getting crio: %v", err)
	}

	portMap := nat.PortMap{}

	utils.AppendIfExplicit(portMap, utils.PortSSH, cmd.Flags(), "ssh-port")
	utils.AppendIfExplicit(portMap, utils.PortVNC, cmd.Flags(), "vnc-port")

	qemuArgs, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return fmt.Errorf("failed getting qemu-args: %v", err)
	}

	cpu, err := cmd.Flags().GetUint("cpu")
	if err != nil {
		return fmt.Errorf("failed getting cpu: %v", err)
	}

	k8sVersion, err := cmd.Flags().GetString("k8s-version")
	if err != nil {
		return fmt.Errorf("failed getting k8s-version: %v", err)
	}

	base := args[0]

	cli, err := client.NewEnvClient()
	if err != nil {
		return fmt.Errorf("failed creating new client: %v", err)
	}
	ctx := context.Background()

	containers, volumes, done := docker.NewCleanupHandler(cli, cmd.OutOrStderr())

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
		done <- fmt.Errorf("Interrupt received, clean up")
	}()

	defer func() {
		fmt.Println("Cleaning up containers and volumes")
		done <- fmt.Errorf("please clean up")

		// Wait a little to cleanup to finish
		time.Sleep(2 * time.Second)
	}()

	// Pull the base image
	reader, err := cli.ImagePull(ctx, "docker.io/"+base, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	docker.PrintProgress(reader, os.Stdout)

	// Start dnsmasq
	dnsmasq, err := cli.ContainerCreate(ctx, &container.Config{
		Image: base,
		Env: []string{
			fmt.Sprintf("NUM_NODES=1"),
		},
		Cmd: []string{"/bin/bash", "-c", "/dnsmasq.sh"},
		ExposedPorts: nat.PortSet{
			utils.TCPPortOrDie(utils.PortSSH): {},
			utils.TCPPortOrDie(utils.PortVNC): {},
		},
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: randomPorts,
		PortBindings:    portMap,
	}, nil, prefix+"-dnsmasq")
	if err != nil {
		return fmt.Errorf("failed creating dnsmasq container: %v", err)
	}
	containers <- dnsmasq.ID

	if err := cli.ContainerStart(ctx, dnsmasq.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed starting dnsmasq container: %v", err)
	}

	nodeName := nodeNameFromIndex(1)
	nodeNum := fmt.Sprintf("%02d", 1)

	vol, err := cli.VolumeCreate(ctx, volume.VolumesCreateBody{
		Name: fmt.Sprintf("%s-%s", prefix, nodeName),
	})
	if err != nil {
		return fmt.Errorf("failed creating volume: %v", err)
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
	}, nil, prefix+"-"+nodeName)
	if err != nil {
		return fmt.Errorf("failed creating node container: %v", err)
	}
	containers <- node.ID
	if err := cli.ContainerStart(ctx, node.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed starting node container: %v", err)
	}

	fmt.Printf("Archiving scripts at %s\n", scripts)
	archivedScripts, err := archive.Tar(scripts, archive.Uncompressed)
	if err != nil {
		return fmt.Errorf("failed archiving scripts at %s: %v", scripts, err)
	}

	nodeContainer := docker.Container{Cli: cli, Id: nodeContainer(prefix, nodeName)}

	err = nodeContainer.ExecScript("mkdir -p /scripts/", "create /scripts directory at container")
	if err != nil {
		return err
	}

	fmt.Printf("Copying scripts from %s into container\n", scripts)
	err = cli.CopyToContainer(ctx, node.ID, "/scripts/", archivedScripts, types.CopyToContainerOptions{})
	if err != nil {
		return fmt.Errorf("failed copying scripts to container: %v", err)
	}

	err = nodeContainer.ExecScript("while [ ! -f /ssh_ready ] ; do sleep 1; done && ssh.sh  echo VM is up", "wait for VM to start")
	if err != nil {
		return err
	}

	err = nodeContainer.ExecScript("scp -r -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i vagrant.key -P 22 /scripts/manifests/* vagrant@192.168.66.101:/tmp", "copy manifests into the VM")
	if err != nil {
		return err
	}

	err = nodeContainer.ExecScript("test -f /scripts/provision.sh", "checking if provision.sh script exists")
	if err != nil {
		return err
	}

	err = nodeContainer.ExecScript(fmt.Sprintf("ssh.sh sudo version=%s /bin/bash -s %t < /scripts/provision.sh", k8sVersion, crio), fmt.Sprintf("provisioning node %s", nodeName))
	if err != nil {
		return err
	}

	err = nodeContainer.ExecScript("rm /usr/local/bin/ssh.sh /ssh_ready", "removing ssh artifacts")
	if err != nil {
		return err
	}

	go func(id string) {
		cli.ContainerWait(context.Background(), id)
	}(node.ID)

	fmt.Printf("Commit the container %s\n", tag)
	_, err = cli.ContainerCommit(ctx, node.ID, types.ContainerCommitOptions{Changes: []string{"ENV PROVISIONED TRUE"}, Reference: tag})
	if err != nil {
		return fmt.Errorf("failed to commit the provisioned container %s: %v", tag, err)
	}

	return nil
}
