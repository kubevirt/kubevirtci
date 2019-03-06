package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"kubevirt.io/kubevirtci/gocli/docker"
)

func NewDeployCommand() *cobra.Command {

	run := &cobra.Command{
		Use:   "deploy",
		Short: "deploy posts a pod to a cluster and waits until we can connect to it",
		RunE:  deploy,
		Args:  cobra.ExactArgs(1),
	}
	run.Flags().UintP("nodes", "n", 1, "number of cluster nodes to start")
	run.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	run.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	run.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	run.Flags().BoolP("background", "b", false, "go to background after nodes are up")
	run.Flags().BoolP("reverse", "r", false, "revert node startup order")
	run.Flags().Bool("random-ports", true, "expose all ports on random localhost ports")
	run.Flags().String("registry-volume", "", "cache docker registry content in the specified volume")
	run.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	run.Flags().Uint("registry-port", 0, "port on localhost for the docker registry")
	run.Flags().Uint("ocp-port", 0, "port on localhost for the ocp cluster")
	run.Flags().Uint("k8s-port", 0, "port on localhost for the k8s cluster")
	run.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")
	run.Flags().String("nfs-data", "", "path to data which should be exposed via nfs to the nodes")
	run.Flags().String("log-to-dir", "", "enables aggregated cluster logging to the folder")
	return run
}

func deploy(cmd *cobra.Command, args []string) (err error) {

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

	logDir, err := cmd.Flags().GetString("log-to-dir")
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

	b := context.Background()
	ctx, cancel := context.WithCancel(b)

	containers, volumes, done := docker.NewCleanupHandler(cli, cmd.OutOrStderr())

	defer func() {
		done <- err
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
		cancel()
		done <- fmt.Errorf("Interrupt received, clean up")
	}()

	// Pull the cluster image
	err = docker.ImagePull(cli, ctx, "docker.io/"+cluster, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

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
	err = docker.ImagePull(cli, ctx, DockerRegistryImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

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
		Image: DockerRegistryImage,
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
		err = docker.ImagePull(cli, ctx, NFSGaneshaImage, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		// Start the ganesha image
		nfsServer, err := cli.ContainerCreate(ctx, &container.Config{
			Image: NFSGaneshaImage,
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

	if logDir != "" {
		logDir, err := filepath.Abs(logDir)
		if err != nil {
			return err
		}

		if _, err = os.Stat(logDir); os.IsNotExist(err) {
			os.Mkdir(logDir, 0755)
		}

		// Pull the fluent image
		err = docker.ImagePull(cli, ctx, FluentdImage, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		// Start the fluent image
		fluentd, err := cli.ContainerCreate(ctx, &container.Config{
			Image: FluentdImage,
			Cmd: strslice.StrSlice{
				"exec fluentd",
				"-i \"<system>\n log_level debug\n</system>\n<source>\n@type  forward\n@log_level error\nport  24224\n</source>\n<match **>\n@type file\npath /fluentd/log/collected\n</match>\"",
				"-p /fluentd/plugins $FLUENTD_OPT -v",
			},
		}, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: logDir,
					Target: "/fluentd/log/collected",
				},
			},
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
		}, nil, prefix+"-fluentd")
		if err != nil {
			return err
		}
		containers <- fluentd.ID
		if err := cli.ContainerStart(ctx, fluentd.ID, types.ContainerStartOptions{}); err != nil {
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

		if len(qemu_args) > 0 {
			qemu_args = "--qemu-args " + qemu_args
		}

		// for docker we manage the startup ourselves, they should just start if it is their turn,
		// tell them all that they are the first node.
		node, err := cli.ContainerCreate(ctx, &container.Config{
			Hostname: nodeName,
			Image:    cluster,
			Env: []string{
				fmt.Sprintf("NODE_NUM=%s", nodeNum),
				"FIRST_NODE=true",
			},
			Volumes: map[string]struct{}{
				"/var/run/disk/": {},
			},
			Cmd: []string{"/bin/bash", "-c", fmt.Sprintf("/vm.sh -n /var/run/disk/disk.qcow2 --memory %s --cpu %s %s", memory, strconv.Itoa(int(cpu)), qemu_args)},
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

		// Wait for vm to be started and provisioned
		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{fmt.Sprintf("/bin/bash", "-c", "while [ ! -f /shared/%s.ready ] ; do sleep 1; done", nodeName)}, os.Stdout)
		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("provisioning node %s failed", nodeName)
		}

		go func(id string) {
			cli.ContainerWait(ctx, id)
			wg.Done()
		}(node.ID)
	}

	// If logging is enabled, deploy the default fluent logging
	if logDir != "" {
		nodeName := nodeNameFromIndex(1)
		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{
			"/bin/bash",
			"-c",
			"ssh.sh sudo /bin/bash < /scripts/logging.sh",
		}, os.Stdout)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("provisioning logging failed")
		}
	}

	// If background flag was specified, we don't want to clean up if we reach that state
	if !background {
		wg.Wait()
		done <- fmt.Errorf("Done. please clean up")
	}

	return nil
}
