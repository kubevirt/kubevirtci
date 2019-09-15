package okd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/spf13/cobra"

	"golang.org/x/net/context"

	"kubevirt.io/kubevirtci/gocli/docker"
)

type copyConfig struct {
	srcPath   string
	dstPath   string
	container string
}

// NewProvisionCommand provision the OKD cluster with one master and one worker
func NewProvisionCommand() *cobra.Command {
	provision := &cobra.Command{
		Use:   "okd",
		Short: "provision okd command will provision new OKD cluster",
		RunE:  provision,
		Args:  cobra.ExactArgs(1),
	}

	provision.Flags().String("dir-hacks", "", "directory with installer hack that should be copied to the container")
	provision.Flags().String("dir-manifests", "", "directory with additional manifests that should be installed")
	provision.Flags().String("dir-scripts", "", "directory with scripts that should be copied to the container")
	provision.Flags().String("master-memory", "8192", "amount of RAM in MB on the master")
	provision.Flags().String("master-cpu", "4", "number of CPU cores on the master")
	provision.Flags().String("workers-memory", "4096", "amount of RAM in MB per worker")
	provision.Flags().String("workers-cpu", "2", "number of CPU per worker")
	provision.Flags().String("installer-pull-token-file", "", "path to the file that contains installer pull token")
	provision.Flags().String("installer-repo-tag", "", "installer repository tag that you want to compile from")
	provision.Flags().String("installer-release-image", "", "the OKD release image that you want to use")
	provision.Flags().Bool("network-operator", false, "install network operator")

	return provision
}

func provision(cmd *cobra.Command, args []string) error {
	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	dirHacks, err := cmd.Flags().GetString("dir-hacks")
	if err != nil {
		return err
	}

	dirManifests, err := cmd.Flags().GetString("dir-manifests")
	if err != nil {
		return err
	}

	dirScripts, err := cmd.Flags().GetString("dir-scripts")
	if err != nil {
		return err
	}

	if dirScripts == "" {
		return fmt.Errorf("you should provide the directory with scripts")
	}

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

	pullTokenFile, err := cmd.Flags().GetString("installer-pull-token-file")
	if err != nil {
		return err
	}

	if pullTokenFile == "" {
		return fmt.Errorf("you should provide the installer pull token file")
	}

	installerTag, err := cmd.Flags().GetString("installer-repo-tag")
	if err != nil {
		return err
	}

	if installerTag == "" {
		return fmt.Errorf("you should provide the installer tag")
	}
	envs = append(envs, fmt.Sprintf("INSTALLER_TAG=%s", installerTag))

	installerReleaseImage, err := cmd.Flags().GetString("installer-release-image")
	if err != nil {
		return err
	}

	if installerReleaseImage != "" {
		envs = append(envs, fmt.Sprintf("INSTALLER_RELEASE_IMAGE=%s", installerReleaseImage))
	}

	networkOperator, err := cmd.Flags().GetBool("network-operator")
	if err != nil {
		return err
	}
	envs = append(envs, fmt.Sprintf("NETWORK_OPERATOR=%v", networkOperator))

	base := args[0]

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

	// Pull the base image
	baseImage := "docker.io/" + base
	fmt.Printf("Download the image %s\n", baseImage)
	reader, err := cli.ImagePull(ctx, baseImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	docker.PrintProgress(reader, os.Stdout)

	clusterContainerName := prefix + "-cluster"
	// Start cluster container
	cluster, err := cli.ContainerCreate(ctx, &container.Config{
		Image: base,
		Env:   envs,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: pullTokenFile,
				Target: "/etc/installer/token",
			},
		},
		Privileged: true,
	}, nil, clusterContainerName)
	if err != nil {
		return err
	}
	containers <- cluster.ID

	fmt.Printf("Start the container %s\n", clusterContainerName)
	if err := cli.ContainerStart(ctx, cluster.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Copy hacks directory to the container
	if dirHacks != "" {
		fmt.Printf("Copy hacks directory to the container %s\n", clusterContainerName)
		config := &copyConfig{
			srcPath:   dirHacks,
			dstPath:   "/",
			container: cluster.ID,
		}
		err = copyToContainer(ctx, cli, config)
		if err != nil {
			return err
		}
	}

	// Copy manifests directory to the container
	if dirManifests != "" {
		fmt.Printf("Copy manifests directory to the container %s\n", clusterContainerName)
		config := &copyConfig{
			srcPath:   dirManifests,
			dstPath:   "/",
			container: cluster.ID,
		}
		err = copyToContainer(ctx, cli, config)
		if err != nil {
			return err
		}
	}

	// Copy scripts directory to the container
	fmt.Printf("Copy scripts directory to the container %s\n", clusterContainerName)
	config := &copyConfig{
		srcPath:   dirScripts,
		dstPath:   "/",
		container: cluster.ID,
	}
	err = copyToContainer(ctx, cli, config)
	if err != nil {
		return err
	}

	// Run provision script
	fmt.Printf("Run provision script\n")
	success, err := docker.Exec(cli, clusterContainerName, []string{"/bin/bash", "-c", "/scripts/provision.sh"}, os.Stdout)
	if err != nil {
		return err
	}

	if !success {
		return fmt.Errorf("failed to provision OKD cluster under the container %s", clusterContainerName)
	}

	fmt.Printf("Commit the container %s\n", clusterContainerName)
	_, err = cli.ContainerCommit(ctx, clusterContainerName, types.ContainerCommitOptions{Reference: "kubevirtci/" + prefix})
	if err != nil {
		return fmt.Errorf("failed to commit the provisioned container %s: %v", clusterContainerName, err)
	}

	done <- fmt.Errorf("Done. Cleanup")

	return nil
}

func copyToContainer(ctx context.Context, cli *client.Client, config *copyConfig) error {
	dstInfo := archive.CopyInfo{
		Exists: true,
		IsDir:  true,
		Path:   config.dstPath,
	}

	srcInfo, err := archive.CopyInfoSourcePath(config.srcPath, true)
	if err != nil {
		return err
	}

	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return err
	}
	defer srcArchive.Close()

	dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
	if err != nil {
		return err
	}
	defer preparedArchive.Close()

	return cli.CopyToContainer(ctx, config.container, dstDir, preparedArchive, types.CopyToContainerOptions{})
}
