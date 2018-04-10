package cmd

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/rmohr/cli/docker"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"io"
	"os"
	"strconv"
)

func NewRunCommand() *cobra.Command {

	run := &cobra.Command{
		Use:   "run",
		Short: "run starts a given cluster",
		RunE:  run,
		Args:  cobra.ExactArgs(1),
	}
	run.Flags().UintP("nodes", "n", 1, "number of cluster nodes to start")
	run.Flags().StringP("memory", "m", "3096M", "amount of ram per nodeName")
	run.Flags().UintP("cpu", "c", 2, "number of cpu cores per nodeName")
	run.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	run.Flags().BoolP("background", "b", false, "go to background after nodes are up")
	run.Flags().BoolP("reverse", "r", false, "revert nodeName startup order")
	run.Flags().String("registry-volume", "", "cache docker registry content in the specified volume")
	run.Flags().String("nfs-data", "", "path to data which should be exposed via nfs to the nodes")
	return run
}

func run(cmd *cobra.Command, args []string) error {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	nodes, err := cmd.Flags().GetUint("nodes")
	if err != nil {
		return err
	}

	memory, err := cmd.Flags().GetString("memory")
	if err != nil {
		return err
	}

	qemu_args, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}

	cpu, err := cmd.Flags().GetUint("cpu")
	if err != nil {
		return err
	}

	cluster := args[0]

	background, err := cmd.Flags().GetBool("background")
	if err != nil {
		return err
	}

	for x := 0; x < int(nodes); x++ {

	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	createdContainers := []string{}
	createdVolumes := []string{}
	ctx := context.Background()

	defer func() {
		for _, c := range createdContainers {
			err := cli.ContainerRemove(ctx, c, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "%v\n", err)
			}
		}

		for _, v := range createdVolumes {
			err := cli.VolumeRemove(ctx, v, true)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "%v\n", err)
			}
		}
	}()

	// Pull the cluster image
	reader, err := cli.ImagePull(ctx, "docker.io/"+cluster, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	// Start dnsmasq
	dnsmasq, err := cli.ContainerCreate(ctx, &container.Config{
		Image: cluster,
		Env: []string{
			fmt.Sprintf("NUM_NODES=%d", nodes),
		},
		Cmd: []string{"/bin/bash", "-c", "/dnsmasq.sh"},
	}, &container.HostConfig{
		Privileged: true,
		ExtraHosts: []string{
			"nfs:192.168.66.2",
			"registry:192.168.66.2",
		},
	}, nil, prefix+"-dnsmasq")
	if err != nil {
		return err
	}
	createdContainers = append(createdContainers, dnsmasq.ID)
	if err := cli.ContainerStart(ctx, dnsmasq.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// start one vm after each other
	for x := 0; x < int(nodes); x++ {
		vol, err := cli.VolumeCreate(ctx, volume.VolumesCreateBody{
			Name: fmt.Sprintf("%s-%s", prefix, nodeName(x+1)),
		})
		if err != nil {
			return err
		}
		createdVolumes = append(createdVolumes, vol.Name)
		node, err := cli.ContainerCreate(ctx, &container.Config{
			Image: cluster,
			Env: []string{
				fmt.Sprintf("NUM_NODES=%d", nodes),
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
		}, nil, prefix+"-"+nodeName(x+1))
		if err != nil {
			return err
		}
		createdContainers = append(createdContainers, node.ID)
		if err := cli.ContainerStart(ctx, node.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}

		// Wait for vm start
		success, err := docker.Exec(cli, nodeContainer(prefix, x+1), []string{"/bin/bash", "-c", "while [ ! -f /usr/local/bin/ssh.sh ] ; do sleep 1; done"}, os.Stdout)
		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("checking for ssh.sh script for node %s failed", nodeName(x+1))
		}

		//check if we have a special provision script
		success, err = docker.Exec(cli, nodeContainer(prefix, x+1), []string{"/bin/bash", "-c", fmt.Sprintf("test -f /scripts/%s.sh", nodeName(x+1))}, os.Stdout)
		if err != nil {
			return fmt.Errorf("checking for matching provision script for node %s failed", nodeName(x+1))
		}

		if success {
			success, err = docker.Exec(cli, nodeContainer(prefix, x+1), []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo /bin/bash < /scripts/%s.sh", nodeName(x+1))}, os.Stdout)
		} else {
			success, err = docker.Exec(cli, nodeContainer(prefix, x+1), []string{"/bin/bash", "-c", "ssh.sh sudo /bin/bash < /scripts/nodes.sh"}, os.Stdout)
		}

		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("provisioning node %s failed", nodeName(x+1))
		}
	}

	// If background flag was specified, we don't want to clean up if we reach that state
	if background {
		createdContainers = []string{}
		createdVolumes = []string{}
	}

	return nil
}

func nodeName(x int) string {
	return fmt.Sprintf("node%02d", x)
}

func nodeContainer(prefix string, x int) string {
	n := nodeName(x)
	return prefix + "-" + n
}
