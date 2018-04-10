package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func NewRootCommand() *cobra.Command {

	root := &cobra.Command{
		Use:   "cli",
		Short: "cli helps you creating ephemeral kubernetes and openshift clusters for testing",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStderr(), cmd.UsageString())
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringP("prefix", "p", "kubevirt", "Prefix to identify docker containers")

	root.AddCommand(
		NewPortCommand(),
		NewRemoveCommand(),
		NewRunCommand(),
	)

	return root

}

func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
