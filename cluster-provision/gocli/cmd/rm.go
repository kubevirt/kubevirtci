package cmd

import (
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

// NewRemoveCommand returns command to remove the cluster
func NewRemoveCommand() *cobra.Command {

	port := &cobra.Command{
		Use:   "rm",
		Short: "rm deletes all traces of a cluster",
		RunE:  rm,
		Args:  cobra.NoArgs,
	}
	return port
}

func rm(cmd *cobra.Command, _ []string) error {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	containers, err := docker.GetPrefixedContainers(cli, prefix+"-")
	if err != nil {
		return err
	}

	var dnsmasq *types.Container
nodnsmasq:
	for i, c := range containers {
		for _, name := range c.Names {
			if strings.HasSuffix(name, "dnsmasq") {
				dnsmasq = &containers[i]
				continue nodnsmasq
			}
		}
		err := cli.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			return err
		}
	}

	// delete dnsmasq at the end since other containers rely on ints network namespace
	if dnsmasq != nil {
		err := cli.ContainerRemove(context.Background(), dnsmasq.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			return err
		}
	}

	volumes, err := docker.GetPrefixedVolumes(cli, prefix)
	if err != nil {
		return err
	}

	for _, v := range volumes {
		err := cli.VolumeRemove(context.Background(), v.Name, true)
		if err != nil {
			return err
		}
	}

	return nil
}
