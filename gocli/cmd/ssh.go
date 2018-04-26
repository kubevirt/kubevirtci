package cmd

import (
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"kubevirt.io/kubevirtci/gocli/docker"
	"os"
)

func NewSSHCommand() *cobra.Command {

	ssh := &cobra.Command{
		Use:   "ssh",
		Short: "ssh into a node",
		RunE:  ssh,
		Args:  cobra.MinimumNArgs(1),
	}
	return ssh
}

func ssh(cmd *cobra.Command, args []string) error {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	node := args[0]

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	// TODO we can do the ssh session with the native golang client
	exitCode, err := docker.Terminal(cli, prefix+"-"+node, append([]string{"ssh.sh"}, args[1:]...), os.Stdout)
	if err != nil {
		return err
	}
	os.Exit(exitCode)
	return nil
}
