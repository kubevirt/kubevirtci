package cmd

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/rmohr/cli/docker"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

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

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	containers, err := docker.GetPrefixedContainers(cli, prefix+"-")
	if err != nil {
		return err
	}

	for _, c := range containers {
		err := cli.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{Force: true})
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
