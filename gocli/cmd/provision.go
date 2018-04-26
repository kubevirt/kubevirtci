package cmd

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"kubevirt.io/kubevirtci/gocli/docker"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"io"
	"os"
	"os/signal"
	"strconv"
)

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

	random_ports, err := cmd.Flags().GetBool("random-ports")
	if err != nil {
		return err
	}

	portMap := nat.PortMap{}

	appendIfExplicit(portMap, PORT_SSH, cmd.Flags(), "ssh-port")
	appendIfExplicit(portMap, PORT_VNC, cmd.Flags(), "vnc-port")

	qemu_args, err := cmd.Flags().GetString("qemu-args")
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
	io.Copy(os.Stdout, reader)

	// Start dnsmasq
	dnsmasq, err := cli.ContainerCreate(ctx, &container.Config{
		Image: base,
		Env: []string{
			fmt.Sprintf("NUM_NODES=1"),
		},
		Cmd: []string{"/bin/bash", "-c", "/dnsmasq.sh"},
		ExposedPorts: nat.PortSet{
			tcpPortOrDie(PORT_SSH): {},
			tcpPortOrDie(PORT_VNC): {},
		},
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: random_ports,
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
	node, err := cli.ContainerCreate(ctx, &container.Config{
		Image: base,
		Env: []string{
			fmt.Sprintf("NODE_NUM=%s", nodeNum),
		},
		Volumes: map[string]struct{}{
			"/var/run/disk/": {},
		},
		Cmd: []string{"/bin/bash", "-c", "/vm.sh", "-n", "/var/run/disk/disk.qcow2", "--memory", memory, "--cpu", strconv.Itoa(int(cpu)), "--qemu-args", qemu_args},
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
	success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "while [ ! -f /usr/local/bin/ssh.sh ] ; do sleep 1; done"}, os.Stdout)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("checking for ssh.sh script for node %s failed", nodeName)
	}

	//check if we have a special provision script
	success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("test -f /scripts/provision.sh", nodeName)}, os.Stdout)
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

	go func(id string) {
		cli.ContainerWait(context.Background(), id)
	}(node.ID)

	return nil
}
