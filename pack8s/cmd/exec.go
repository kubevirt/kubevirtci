package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/fromanirh/pack8s/cmd/cmdutil"
)

type execOptions struct {
	commands []string
}

var execOpt execOptions

// NewExecCommand runs given command inside container
func NewExecCommand() *cobra.Command {
	exec := &cobra.Command{
		Use:   "exec",
		Short: "exec runs given command in container",
		RunE:  execCommand,
		Args:  cobra.MinimumNArgs(2),
	}

	return exec
}

func execCommand(cmd *cobra.Command, args []string) error {
	containerID := args[0]
	command := args[1:]

	cOpts, err := cmdutil.GetCommonOpts(cmd)
	if err != nil {
		return err
	}

	hnd, _, err := cOpts.GetHandle()
	if err != nil {
		return err
	}

	return hnd.Exec(containerID, command, os.Stdout)
}
