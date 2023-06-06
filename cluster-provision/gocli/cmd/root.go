package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	if os.Getenv("DOCKER_API_VERSION") == "" {
		os.Setenv("DOCKER_API_VERSION", "1.24")
	}
}

// NewRootCommand returns entrypoint command to interact with all other commands
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
		NewProvisionCommand(),
		NewRemoveCommand(),
		NewRunCommand(),
		NewSSHCommand(),
		NewSCPCommand(),
	)

	return root

}

// Execute executes root command
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
