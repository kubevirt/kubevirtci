package cmd

import (
	"os"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

// NewSSHCommand returns command to SSH to the cluster node
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

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	// TODO we can do the ssh session with the native golang client
	container := prefix + "-" + node
	ssh_command := append([]string{"ssh.sh"}, args[1:]...)
	file := os.Stdout
	if terminal.IsTerminal(int(file.Fd())) {
		exitCode, err := docker.Terminal(cli, container, ssh_command, file)
		if err != nil {
			return err
		}
		os.Exit(exitCode)
	} else {
		execExitCodeIsZero, err := docker.Exec(cli, container, ssh_command, file)
		if err != nil {
			return err
		}
		exitCode := 0
		if !execExitCodeIsZero {
			exitCode = 1
		}
		os.Exit(exitCode)
	}
	return nil
}
