package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/okd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

// NewProvisionCommand provision given cluster
func NewProvisionCommand() *cobra.Command {

	provision := &cobra.Command{
		Use:   "provision",
		Short: "provision starts a given cluster",
		RunE:  provision,
		Args:  cobra.ExactArgs(2),
	}
	provision.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	provision.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	provision.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	provision.Flags().String("scripts", "", "location for the provision and run scripts")
	provision.Flags().Bool("random-ports", false, "expose all ports on random localhost ports")
	provision.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	provision.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")

	provision.AddCommand(
		okd.NewProvisionCommand(),
	)

	return provision
}

func provision(cmd *cobra.Command, args []string) error {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	scripts, err := cmd.Flags().GetString("scripts")
	if err != nil {
		return err
	}

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
	utils.AppendIfExplicit(portMap, utils.VNCPortStartRange+1, cmd.Flags(), "vnc-port")

	qemuArgs, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}

	cpu, err := cmd.Flags().GetUint("cpu")
	if err != nil {
		return err
	}

	base := args[0]
	//target := args[1]

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	containers, volumes, done := docker.NewCleanupHandler(cli, cmd.OutOrStderr())

	defer func() {
		done <- fmt.Errorf("please clean up")
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
		done <- fmt.Errorf("Interrupt received, clean up")
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
			utils.TCPPortOrDie(utils.PortSSH):               {},
			utils.TCPPortOrDie(utils.VNCPortStartRange + 1): {},
		},
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: randomPorts,
		PortBindings:    portMap,
	}, nil, prefix+"-dnsmasq")
	if err != nil {
		return err
	}
	containers <- dnsmasq.ID
	if err := cli.ContainerStart(ctx, dnsmasq.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	nodeName := nodeNameFromIndex(1)
	nodeNum := fmt.Sprintf("%02d", 1)

	vol, err := cli.VolumeCreate(ctx, volume.VolumesCreateBody{
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
		Cmd: []string{"/bin/bash", "-c", fmt.Sprintf("/vm.sh -n /var/run/disk/disk.qcow2 --memory %s --cpu %s %s", memory, strconv.Itoa(int(cpu)), qemuArgs)},
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
		return err
	}
	containers <- node.ID
	if err := cli.ContainerStart(ctx, node.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Copy scripts
	fmt.Println(scripts)

	// Wait for vm start
	success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "while [ ! -f /ssh_ready ] ; do sleep 1; done"}, os.Stdout)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("checking for ssh.sh script for node %s failed", nodeName)
	}

	//check if we have a special provision script
	success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "test -f /scripts/provision.sh"}, os.Stdout)
	if err != nil {
		return fmt.Errorf("checking for a provision script failed failed: %v", err)
	}

	success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "ssh.sh sudo /bin/bash < /scripts/provision.sh"}, os.Stdout)

	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("provisioning node %s failed", nodeName)
	}

	success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "rm /usr/local/bin/ssh.sh"}, os.Stdout)
	success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "rm /ssh_ready"}, os.Stdout)

	go func(id string) {
		cli.ContainerWait(context.Background(), id)
	}(node.ID)

	return nil
}
