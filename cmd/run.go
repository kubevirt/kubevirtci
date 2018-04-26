package cmd

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rmohr/cli/docker"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/net/context"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
)

func NewRunCommand() *cobra.Command {

	run := &cobra.Command{
		Use:   "run",
		Short: "run starts a given cluster",
		RunE:  run,
		Args:  cobra.ExactArgs(1),
	}
	run.Flags().UintP("nodes", "n", 1, "number of cluster nodes to start")
	run.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	run.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	run.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	run.Flags().BoolP("background", "b", false, "go to background after nodes are up")
	run.Flags().BoolP("reverse", "r", false, "revert node startup order")
	run.Flags().Bool("random-ports", false, "expose all ports on random localhost ports")
	run.Flags().String("registry-volume", "", "cache docker registry content in the specified volume")
	run.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	run.Flags().Uint("registry-port", 0, "port on localhost for the docker registry")
	run.Flags().Uint("ocp-port", 0, "port on localhost for the ocp cluster")
	run.Flags().Uint("k8s-port", 0, "port on localhost for the k8s cluster")
	run.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")
	run.Flags().String("nfs-data", "", "path to data which should be exposed via nfs to the nodes")
	return run
}

func run(cmd *cobra.Command, args []string) (err error) {

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

	reverse, err := cmd.Flags().GetBool("reverse")
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
	appendIfExplicit(portMap, PORT_K8S, cmd.Flags(), "k8s-port")
	appendIfExplicit(portMap, PORT_OCP, cmd.Flags(), "ocp-port")
	appendIfExplicit(portMap, PORT_REGISTRY, cmd.Flags(), "registry-port")

	qemu_args, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}

	cpu, err := cmd.Flags().GetUint("cpu")
	if err != nil {
		return err
	}

	registry_volume, err := cmd.Flags().GetString("registry-volume")
	if err != nil {
		return err
	}

	nfs_data, err := cmd.Flags().GetString("nfs-data")
	if err != nil {
		return err
	}

	cluster := args[0]

	background, err := cmd.Flags().GetBool("background")
	if err != nil {
		return err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	containers := make(chan string)
	volumes := make(chan string)
	done := make(chan error)

	go func() {
		createdContainers := []string{}
		createdVolumes := []string{}

		for {
			select {
			case container := <-containers:
				createdContainers = append(createdContainers, container)
			case volume := <-volumes:
				createdVolumes = append(createdVolumes, volume)
			case err := <-done:
				if err != nil {
					for _, c := range createdContainers {
						err := cli.ContainerRemove(ctx, c, types.ContainerRemoveOptions{Force: true})
						fmt.Printf("container: %v\n", c)
						if err != nil {
							fmt.Fprintf(cmd.OutOrStderr(), "%v\n", err)
						}
					}

					for _, v := range createdVolumes {
						err := cli.VolumeRemove(ctx, v, true)
						fmt.Printf("volume: %v\n", v)
						if err != nil {
							fmt.Fprintf(cmd.OutOrStderr(), "%v\n", err)
						}
					}
				}
			}
		}
	}()

	defer func() {
		fmt.Printf("error is %v\n", err)
		done <- err
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
		done <- fmt.Errorf("Interrupt received, clean up")
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
		ExposedPorts: nat.PortSet{
			tcpPortOrDie(PORT_SSH):      {},
			tcpPortOrDie(PORT_REGISTRY): {},
			tcpPortOrDie(PORT_OCP):      {},
			tcpPortOrDie(PORT_K8S):      {},
			tcpPortOrDie(PORT_VNC):      {},
		},
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: random_ports,
		PortBindings:    portMap,
		ExtraHosts: []string{
			"nfs:192.168.66.2",
			"registry:192.168.66.2",
		},
	}, nil, prefix+"-dnsmasq")
	if err != nil {
		return err
	}
	containers <- dnsmasq.ID
	if err := cli.ContainerStart(ctx, dnsmasq.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Pull the registry image
	reader, err = cli.ImagePull(ctx, "docker.io/library/registry:2", types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)

	// Create registry volume
	var registryMounts []mount.Mount
	if registry_volume != "" {

		vol, err := cli.VolumeCreate(ctx, volume.VolumesCreateBody{
			Name: fmt.Sprintf("%s-%s", prefix, "registry"),
		})
		if err != nil {
			return err
		}
		registryMounts = []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: vol.Name,
				Target: "/var/lib/registry",
			},
		}
	}

	// Start registry
	registry, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "registry:2",
	}, &container.HostConfig{
		Mounts:      registryMounts,
		Privileged:  true, // fixme we just need proper selinux volume labeling
		NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
	}, nil, prefix+"-registry")
	if err != nil {
		return err
	}
	containers <- registry.ID
	if err := cli.ContainerStart(ctx, registry.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	if nfs_data != "" {
		nfs_data, err := filepath.Abs(nfs_data)
		if err != nil {
			return err
		}
		// Pull the ganesha image
		reader, err = cli.ImagePull(ctx, "docker.io/janeczku/nfs-ganesha", types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, reader)

		// Start the ganesha image
		nfsServer, err := cli.ContainerCreate(ctx, &container.Config{
			Image: "janeczku/nfs-ganesha",
		}, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: nfs_data,
					Target: "/data/nfs",
				},
			},
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
		}, nil, prefix+"-nfs-ganesha")
		if err != nil {
			return err
		}
		containers <- nfsServer.ID
		if err := cli.ContainerStart(ctx, nfsServer.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(int(nodes))
	// start one vm after each other
	for x := 0; x < int(nodes); x++ {

		nodeName := nodeNameFromIndex(x + 1)
		nodeNum := fmt.Sprintf("%02d", x+1)
		if reverse {
			nodeName = nodeNameFromIndex((int(nodes) - x))
			nodeNum = fmt.Sprintf("%02d", (int(nodes) - x))
		}

		vol, err := cli.VolumeCreate(ctx, volume.VolumesCreateBody{
			Name: fmt.Sprintf("%s-%s", prefix, nodeName),
		})
		if err != nil {
			return err
		}
		volumes <- vol.Name
		node, err := cli.ContainerCreate(ctx, &container.Config{
			Image: cluster,
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

		// Wait for vm start
		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "while [ ! -f /usr/local/bin/ssh.sh ] ; do sleep 1; done"}, os.Stdout)
		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("checking for ssh.sh script for node %s failed", nodeName)
		}

		//check if we have a special provision script
		success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("test -f /scripts/%s.sh", nodeName)}, os.Stdout)
		if err != nil {
			return fmt.Errorf("checking for matching provision script for node %s failed", nodeName)
		}

		if success {
			success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo /bin/bash < /scripts/%s.sh", nodeName)}, os.Stdout)
		} else {
			success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "ssh.sh sudo /bin/bash < /scripts/nodes.sh"}, os.Stdout)
		}

		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("provisioning node %s failed", nodeName)
		}

		go func(id string) {
			cli.ContainerWait(context.Background(), id)
			wg.Done()
		}(node.ID)
	}

	// If background flag was specified, we don't want to clean up if we reach that state
	if !background {
		wg.Wait()
		done <- fmt.Errorf("Done. please clean up")
	}

	return nil
}

func nodeNameFromIndex(x int) string {
	return fmt.Sprintf("node%02d", x)
}

func nodeContainer(prefix string, node string) string {
	return prefix + "-" + node
}

func appendIfExplicit(ports nat.PortMap, exposedPort int, flagSet *pflag.FlagSet, flagName string) error {
	flag := flagSet.Lookup(flagName)
	if flag != nil && flag.Changed {
		publicPort, err := flagSet.GetUint(flagName)
		if err != nil {
			return err
		}
		ports[tcpPortOrDie(exposedPort)] = []nat.PortBinding{{"127.0.0.1", strconv.Itoa(int(publicPort))}}
	}
	return nil
}
