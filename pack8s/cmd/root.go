package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/fromanirh/pack8s/internal/pkg/podman"

	"github.com/fromanirh/pack8s/cmd/cmdutil"
)

// NewRootCommand returns entrypoint command to interact with all other commands
func NewRootCommand() *cobra.Command {

	root := &cobra.Command{
		Use:   "pack8s",
		Short: "pack8s helps you creating ephemeral kubernetes and openshift clusters for testing",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStderr(), cmd.UsageString())
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmdutil.AddCommonOpts(root)

	root.AddCommand(
		NewPortCommand(),
		NewPullCommand(),
		NewRemoveCommand(),
		NewRunCommand(),
		NewSCPCommand(),
		NewSSHCommand(),
		NewShowCommand(),
		NewPruneVolumesCommand(),
		NewExecCommand(),
		NewVersionCommand(),
	)

	return root

}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Println(podman.SprintError("pack8s", err)) //XXX specific method
		os.Exit(1)
	}
}
