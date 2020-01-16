package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/fromanirh/pack8s/cmd/cmdutil"
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
	cOpts, err := cmdutil.GetCommonOpts(cmd)
	if err != nil {
		return err
	}

	node := args[0]

	hnd, _, err := cOpts.GetHandle()
	if err != nil {
		return err
	}

	container := cOpts.Prefix + "-" + node
	sshCommand := append([]string{"ssh.sh"}, args[1:]...)

	err = hnd.Exec(container, sshCommand, os.Stdout)

	return err
}
