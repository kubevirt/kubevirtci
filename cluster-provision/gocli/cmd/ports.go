package cmd

import (
	"context"
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

Known port names are 'ssh', 'registry', 'ocp', 'k8s', 'prometheus' and 'grafana'.
`,
		RunE: ports,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one port name can be specified at once")
			}

			if len(args) == 1 {
				switch args[0] {
				case utils.PortNameSSH, utils.PortNameSSHWorker, utils.PortNameAPI, utils.PortNameOCP, utils.PortNameOCPConsole, utils.PortNameRegistry, utils.PortNameVNC, utils.PortNameHTTP, utils.PortNameHTTPS, utils.PortNamePrometheus, utils.PortNameGrafana, utils.PortNameUploadProxy, utils.PortNameDNS:
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

	cli, err := client.NewClientWithOpts(client.FromEnv)
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
	container, err := cli.ContainerInspect(context.Background(), containers[0].ID)
	if err != nil {
		return err
	}

	if portName != "" {
		err = nil
		switch portName {
		case utils.PortNameSSH:
			err = utils.PrintPublicPort(utils.PortSSH, container.NetworkSettings.Ports)
		case utils.PortNameSSHWorker:
			err = utils.PrintPublicPort(utils.PortSSHWorker, container.NetworkSettings.Ports)
		case utils.PortNameAPI:
			err = utils.PrintPublicPort(utils.PortAPI, container.NetworkSettings.Ports)
		case utils.PortNameRegistry:
			err = utils.PrintPublicPort(utils.PortRegistry, container.NetworkSettings.Ports)
		case utils.PortNameOCP:
			err = utils.PrintPublicPort(utils.PortOCP, container.NetworkSettings.Ports)
		case utils.PortNameOCPConsole:
			err = utils.PrintPublicPort(utils.PortOCPConsole, container.NetworkSettings.Ports)
		case utils.PortNameVNC:
			err = utils.PrintPublicPort(utils.PortVNC, container.NetworkSettings.Ports)
		case utils.PortNameHTTP:
			err = utils.PrintPublicPort(utils.PortHTTP, container.NetworkSettings.Ports)
		case utils.PortNameHTTPS:
			err = utils.PrintPublicPort(utils.PortHTTPS, container.NetworkSettings.Ports)
		case utils.PortNamePrometheus:
			err = utils.PrintPublicPort(utils.PortPrometheus, container.NetworkSettings.Ports)
		case utils.PortNameGrafana:
			err = utils.PrintPublicPort(utils.PortGrafana, container.NetworkSettings.Ports)
		case utils.PortNameUploadProxy:
			err = utils.PrintPublicPort(utils.PortUploadProxy, container.NetworkSettings.Ports)
		case utils.PortNameDNS:
			err = utils.PrintPublicPort(utils.PortDNS, container.NetworkSettings.Ports)
		}

		if err != nil {
			return err
		}

	} else {
		for k, pp := range container.NetworkSettings.Ports {
			for _, p := range pp {
				fmt.Printf("%s -> %s:%s\n", k, p.HostIP, p.HostPort)
			}
		}
	}

	return nil
}
