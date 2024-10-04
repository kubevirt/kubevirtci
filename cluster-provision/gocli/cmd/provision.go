package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/providers"
)

const (
	baseLinuxPhase = "quay.io/kubevirtci/centos9"
	baseK8sPhase   = "quay.io/kubevirtci/centos9:2408130400-bb670376"
)

var versionMap = map[string]string{
	"1.30": "1.30.2",
	"1.29": "1.29.6",
	"1.28": "1.28.11",
	"1.31": "1.31.0",
}

// NewProvisionCommand provision given cluster
func NewProvisionCommand() *cobra.Command {
	provision := &cobra.Command{
		Use:   "provision",
		Short: "provision starts a given cluster",
		RunE:  provisionCluster,
		Args:  cobra.ExactArgs(1),
	}
	provision.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	provision.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	provision.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	provision.Flags().Bool("random-ports", true, "expose all ports on random localhost ports")
	provision.Flags().Bool("slim", false, "create slim provider (uncached images)")
	provision.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	provision.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")
	provision.Flags().String("container-suffix", "", "use additional suffix for the provisioned container image")
	provision.Flags().String("phases", "linux,k8s", "phases to run, possible values: linux,k8s linux k8s")
	provision.Flags().StringArray("additional-persistent-kernel-arguments", []string{}, "additional persistent kernel arguments applied after provision")

	return provision
}

func provisionCluster(cmd *cobra.Command, args []string) (retErr error) {
	var base string
	versionNoMinor := args[0]

	v, ok := versionMap[versionNoMinor]
	if !ok {
		return fmt.Errorf("Invalid version passed, exiting!")
	}

	opts := []providers.KubevirtProviderOption{}
	flags := cmd.Flags()
	for flagName, flagConfig := range providers.ProvisionFlagMap {
		switch flagConfig.FlagType {
		case "string":
			flagVal, err := flags.GetString(flagName)
			if err != nil {
				return err
			}
			opts = append(opts, flagConfig.ProviderOptFunc(flagVal))
		case "bool":
			flagVal, err := flags.GetBool(flagName)
			if err != nil {
				return err
			}
			opts = append(opts, flagConfig.ProviderOptFunc(flagVal))

		case "uint":
			flagVal, err := flags.GetUint(flagName)
			if err != nil {
				return err
			}
			opts = append(opts, flagConfig.ProviderOptFunc(flagVal))
		case "uint16":
			flagVal, err := flags.GetUint16(flagName)
			if err != nil {
				return err
			}
			opts = append(opts, flagConfig.ProviderOptFunc(flagVal))
		case "[]string":
			flagVal, err := flags.GetStringArray(flagName)
			if err != nil {
				return err
			}
			opts = append(opts, flagConfig.ProviderOptFunc(flagVal))
		}
	}

	phases, err := cmd.Flags().GetString("phases")
	if err != nil {
		return err
	}

	if strings.Contains(phases, "linux") {
		base = baseLinuxPhase
	} else {
		base = baseK8sPhase
	}

	containerSuffix, err := cmd.Flags().GetString("container-suffix")
	if err != nil {
		return err
	}
	name := filepath.Base(versionNoMinor)
	if len(containerSuffix) > 0 {
		name = fmt.Sprintf("%s-%s", name, containerSuffix)
	}

	portMap := nat.PortMap{}

	utils.AppendTCPIfExplicit(portMap, utils.PortSSH, cmd.Flags(), "ssh-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortVNC, cmd.Flags(), "vnc-port")

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())

	kp := providers.NewKubevirtProvider(versionNoMinor, base, cli, opts)
	err = kp.Provision(ctx, cancel, portMap, v)
	if err != nil {
		return err
	}

	return nil
}
