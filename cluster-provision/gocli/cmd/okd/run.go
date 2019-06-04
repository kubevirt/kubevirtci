package okd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"kubevirt.io/kubevirtci/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/gocli/docker"
)

// NewRunCommand returns command that runs OKD cluster
func NewRunCommand() *cobra.Command {
	run := &cobra.Command{
		Use:   "okd",
		Short: "run OKD cluster",
		RunE:  run,
		Args:  cobra.ExactArgs(1),
	}
	run.Flags().String("master-memory", "12288", "amount of RAM in MB on the master")
	run.Flags().String("master-cpu", "4", "number of CPU cores on the master")
	run.Flags().String("workers", "1", "number of cluster worker nodes to start")
	run.Flags().String("workers-memory", "6144", "amount of RAM in MB per worker")
	run.Flags().String("workers-cpu", "2", "number of CPU per worker")
	run.Flags().String("registry-volume", "", "cache docker registry content in the specified volume")
	run.Flags().String("nfs-data", "", "path to data which should be exposed via nfs to the nodes")
	run.Flags().Uint("registry-port", 0, "port on localhost for the docker registry")
	run.Flags().Uint("ocp-console-port", 0, "port on localhost for the ocp console")
	run.Flags().Uint("k8s-port", 0, "port on localhost for the k8s cluster")
	run.Flags().Uint("ssh-master-port", 0, "port on localhost to ssh to master node")
	run.Flags().Uint("ssh-worker-port", 0, "port on localhost to ssh to worker node")
	run.Flags().Bool("background", false, "go to background after nodes are up")
	run.Flags().Bool("random-ports", true, "expose all ports on random localhost ports")
	return run
}

func run(cmd *cobra.Command, args []string) (err error) {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	// TODO: set number of workers via machine set under the run.sh script
	// workers, err := cmd.Flags().GetString("workers")
	// if err != nil {
	// 	return err
	// }

	envs := []string{}
	masterMemory, err := cmd.Flags().GetString("master-memory")
	if err != nil {
		return err
	}
	envs = append(envs, fmt.Sprintf("MASTER_MEMORY=%s", masterMemory))

	masterCPU, err := cmd.Flags().GetString("master-cpu")
	if err != nil {
		return err
	}
	envs = append(envs, fmt.Sprintf("MASTER_CPU=%s", masterCPU))

	workersMemory, err := cmd.Flags().GetString("workers-memory")
	if err != nil {
		return err
	}
	envs = append(envs, fmt.Sprintf("WORKERS_MEMORY=%s", workersMemory))

	workersCPU, err := cmd.Flags().GetString("workers-cpu")
	if err != nil {
		return err
	}
	envs = append(envs, fmt.Sprintf("WORKERS_CPU=%s", workersCPU))

	randomPorts, err := cmd.Flags().GetBool("random-ports")
	if err != nil {
		return err
	}

	portMap := nat.PortMap{}

	utils.AppendIfExplicit(portMap, utils.PortSSH, cmd.Flags(), "ssh-master-port")
	utils.AppendIfExplicit(portMap, utils.PortSSHWorker, cmd.Flags(), "ssh-worker-port")
	utils.AppendIfExplicit(portMap, utils.PortAPI, cmd.Flags(), "k8s-port")
	utils.AppendIfExplicit(portMap, utils.PortOCPConsole, cmd.Flags(), "ocp-console-port")
	utils.AppendIfExplicit(portMap, utils.PortRegistry, cmd.Flags(), "registry-port")

	registryVol, err := cmd.Flags().GetString("registry-volume")
	if err != nil {
		return err
	}

	nfsData, err := cmd.Flags().GetString("nfs-data")
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

	containers, _, done := docker.NewCleanupHandler(cli, cmd.OutOrStderr())

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
	fmt.Printf("Download the image %s\n", "docker.io/"+cluster)
	err = docker.ImagePull(cli, ctx, "docker.io/"+cluster, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	clusterContainerName := prefix + "-cluster"
	// Start cluster container
	clusterContainer, err := cli.ContainerCreate(ctx, &container.Config{
		Image: cluster,
		Env:   envs,
		ExposedPorts: nat.PortSet{
			utils.TCPPortOrDie(utils.PortSSH):        {},
			utils.TCPPortOrDie(utils.PortSSHWorker):  {},
			utils.TCPPortOrDie(utils.PortRegistry):   {},
			utils.TCPPortOrDie(utils.PortOCPConsole): {},
			utils.TCPPortOrDie(utils.PortAPI):        {},
		},
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: randomPorts,
		PortBindings:    portMap,
	}, nil, clusterContainerName)
	if err != nil {
		return err
	}
	containers <- clusterContainer.ID
	fmt.Printf("Start the container %s\n", clusterContainerName)
	if err := cli.ContainerStart(ctx, clusterContainer.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Pull the registry image
	fmt.Printf("Download the image %s\n", utils.DockerRegistryImage)
	err = docker.ImagePull(cli, ctx, utils.DockerRegistryImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	// Create registry volume
	var registryMounts []mount.Mount
	if registryVol != "" {
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
		Image: utils.DockerRegistryImage,
	}, &container.HostConfig{
		Mounts:      registryMounts,
		Privileged:  true, // fixme we just need proper selinux volume labeling
		NetworkMode: container.NetworkMode("container:" + clusterContainer.ID),
	}, nil, prefix+"-registry")
	if err != nil {
		return err
	}
	containers <- registry.ID
	fmt.Printf("Start the container %s\n", prefix+"-registry")
	if err := cli.ContainerStart(ctx, registry.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	if nfsData != "" {
		nfsData, err := filepath.Abs(nfsData)
		if err != nil {
			return err
		}
		// Pull the ganesha image
		fmt.Printf("Download the image %s\n", utils.NFSGaneshaImage)
		err = docker.ImagePull(cli, ctx, utils.NFSGaneshaImage, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		// Start the ganesha image
		nfsServer, err := cli.ContainerCreate(ctx, &container.Config{
			Image: utils.NFSGaneshaImage,
		}, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: nfsData,
					Target: "/data/nfs",
				},
			},
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + clusterContainer.ID),
		}, nil, prefix+"-nfs-ganesha")
		if err != nil {
			return err
		}
		containers <- nfsServer.ID
		fmt.Printf("Start the container %s\n", prefix+"-nfs-ganesha")
		if err := cli.ContainerStart(ctx, nfsServer.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
	}

	// Run the cluster
	fmt.Printf("Run the cluster\n")
	success, err := docker.Exec(cli, clusterContainerName, []string{"/bin/bash", "-c", "/scripts/run.sh"}, os.Stdout)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("failed to run the OKD cluster under the container %s", clusterContainerName)
	}

	// If background flag was specified, we don't want to clean up if we reach that state
	if !background {
		done <- fmt.Errorf("Done. please clean up")
	}

	return nil
}
