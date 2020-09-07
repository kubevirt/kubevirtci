package cmd

import (
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
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
				case utils.PortNameSSH, utils.PortNameSSHWorker, utils.PortNameAPI, utils.PortNameOCP, utils.PortNameOCPConsole, utils.PortNameRegistry, utils.PortNameVNC, utils.PortNameHTTP, utils.PortNameHTTPS:
					return nil
				default:
					return fmt.Errorf("unknown port name %s", args[0])
				}
			}
			return nil
		},
	}

	port.Flags().String("container-name", "dnsmasq", "the container name to SSH copy from")

	return port
}

func ports(cmd *cobra.Command, args []string) error {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	containerName, err := cmd.Flags().GetString("container-name")
	if err != nil {
		return err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	containers, err := docker.GetPrefixedContainers(cli, prefix+"-"+containerName)
	if err != nil {
		return err
	}

	if len(containers) != 1 {
		return fmt.Errorf("failed to found the container with name %s", prefix+"-"+containerName)
	}

	portName := ""
	if len(args) > 0 {
		portName = args[0]
	}

	if portName != "" {
		err = nil
		switch portName {
		case utils.PortNameSSH:
			err = utils.PrintPublicPort(utils.PortSSH, containers[0].Ports)
		case utils.PortNameSSHWorker:
			err = utils.PrintPublicPort(utils.PortSSHWorker, containers[0].Ports)
		case utils.PortNameAPI:
			err = utils.PrintPublicPort(utils.PortAPI, containers[0].Ports)
		case utils.PortNameRegistry:
			err = utils.PrintPublicPort(utils.PortRegistry, containers[0].Ports)
		case utils.PortNameOCP:
			err = utils.PrintPublicPort(utils.PortOCP, containers[0].Ports)
		case utils.PortNameOCPConsole:
			err = utils.PrintPublicPort(utils.PortOCPConsole, containers[0].Ports)
		case utils.PortNameVNC:
			err = utils.PrintPublicPort(utils.PortVNC, containers[0].Ports)
		case utils.PortNameHTTP:
			err = utils.PrintPublicPort(utils.PortHTTP, containers[0].Ports)
		case utils.PortNameHTTPS:
			err = utils.PrintPublicPort(utils.PortHTTPS, containers[0].Ports)
		}

		if err != nil {
			return err
		}

	} else {
		for _, p := range containers[0].Ports {
			fmt.Printf("%d/%s -> %s:%d\n", p.PrivatePort, p.Type, p.IP, p.PublicPort)
		}
	}

	return nil
}
