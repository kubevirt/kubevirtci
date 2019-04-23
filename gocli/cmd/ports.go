package cmd

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"kubevirt.io/kubevirtci/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/gocli/docker"
)

// NewPortCommand returns new command to expose public ports for the cluster
func NewPortCommand() *cobra.Command {

	port := &cobra.Command{
		Use:   "ports",
		Short: "ports shows exposed ports of the cluster",
		Long: `ports shows exposed ports of the cluster

If no port name is specified, all exposed ports are printed.
If an extra port name is specified, only the exposed port is printed.

Known port names are 'ssh', 'registry', 'ocp' and 'k8s'.
`,
		RunE: ports,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one port name can be specified at once")
			}

			if len(args) == 1 {
				switch args[0] {
				case utils.PortNameSSH, utils.PortNameAPI, utils.PortNameOCP, utils.PortNameRegistry, utils.PortNameVNC:
					return nil
				default:
					return fmt.Errorf("unknown port name %s", args[0])
				}
			}
			return nil
		},
	}
	return port
}

func ports(cmd *cobra.Command, args []string) error {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	container, err := docker.GetDDNSMasqContainer(cli, prefix)
	if err != nil {
		return err
	}

	portName := ""
	if len(args) > 0 {
		portName = args[0]
	}

	if portName != "" {
		err = nil
		switch portName {
		case utils.PortNameSSH:
			err = utils.PrintPublicPort(utils.PortSSH, container.Ports)
		case utils.PortNameAPI:
			err = utils.PrintPublicPort(utils.PortAPI, container.Ports)
		case utils.PortNameRegistry:
			err = utils.PrintPublicPort(utils.PortRegistry, container.Ports)
		case utils.PortNameOCP:
			err = utils.PrintPublicPort(utils.PortOCP, container.Ports)
		case utils.PortNameVNC:
			err = utils.PrintPublicPort(utils.PortVNC, container.Ports)
		}

		if err != nil {
			return err
		}

	} else {
		for _, p := range container.Ports {
			fmt.Printf("%d/%s -> %s:%d\n", p.PrivatePort, p.Type, p.IP, p.PublicPort)
		}
	}

	return nil
}
