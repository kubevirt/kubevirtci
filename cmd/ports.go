package cmd

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rmohr/cli/docker"
	"github.com/spf13/cobra"
	"strconv"
)

const (
	PORT_SSH      = 2201
	PORT_REGISTRY = 5000
	PORT_OCP      = 8443
	PORT_K8S      = 6443
	PORT_VNC      = 5901

	PORT_NAME_SSH      = "ssh"
	PORT_NAME_OCP      = "ocp"
	PORT_NAME_REGISTRY = "registry"
	PORT_NAME_K8S      = "k8s"
	PORT_NAME_VNC      = "vnc"
)

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
				case PORT_NAME_SSH, PORT_NAME_K8S, PORT_NAME_OCP, PORT_NAME_REGISTRY, PORT_NAME_VNC:
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
		case PORT_NAME_SSH:
			err = printPort(PORT_SSH, container.Ports)
		case PORT_NAME_K8S:
			err = printPort(PORT_K8S, container.Ports)
		case PORT_NAME_REGISTRY:
			err = printPort(PORT_REGISTRY, container.Ports)
		case PORT_NAME_OCP:
			err = printPort(PORT_OCP, container.Ports)
		case PORT_NAME_VNC:
			err = printPort(PORT_VNC, container.Ports)
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

func getPort(port uint16, ports []types.Port) (uint16, error) {
	for _, p := range ports {
		if p.PrivatePort == port {
			return p.PublicPort, nil
		}
	}
	return 0, fmt.Errorf("port is not exposed")
}

func printPort(port uint16, ports []types.Port) error {
	p, err := getPort(port, ports)
	if err != nil {
		return err
	}
	fmt.Println(p)
	return nil
}

func tcpPortOrDie(port int) nat.Port {
	p, err := nat.NewPort("tcp", strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	return p
}
