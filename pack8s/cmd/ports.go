package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fromanirh/pack8s/cmd/cmdutil"

	"github.com/fromanirh/pack8s/internal/pkg/ports"
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
		RunE: showPorts,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one port name can be specified at once")
			}

			if len(args) == 1 {
				if !ports.IsKnownPortName(args[0]) {
					return fmt.Errorf("unknown port name %s", args[0])
				}
			}
			return nil
		},
	}

	port.Flags().String("container-name", "dnsmasq", "the container name to SSH copy from")

	return port
}

func showPorts(cmd *cobra.Command, args []string) error {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	cOpts, err := cmdutil.GetCommonOpts(cmd)
	if err != nil {
		return err
	}

	containerName, err := cmd.Flags().GetString("container-name")
	if err != nil {
		return err
	}

	hnd, _, err := cOpts.GetHandle()
	if err != nil {
		return err
	}

	cont, err := hnd.FindPrefixedContainer(prefix + "-" + containerName)
	if err != nil {
		return err
	}

	portName := ""
	if len(args) > 0 {
		portName = args[0]
	}

	if portName != "" {
		port, err := ports.NameToNumber(portName)
		if err != nil {
			return err
		}

		err = ports.PrintPublicPort(port, cont.Ports)
		if err != nil {
			return err
		}
	} else {
		for _, p := range cont.Ports {
			fmt.Printf("%s/%s -> %s:%s\n", p.Container_port, p.Protocol, p.Host_ip, p.Host_port)
		}
	}

	return nil
}
