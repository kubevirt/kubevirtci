package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	kind "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/kindbase"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/sriov"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/vgpu"
)

var kindProvider kind.KindProvider

func NewRunKindCommand() *cobra.Command {
	rk := &cobra.Command{
		Use:   "run-kind",
		Short: "runs a kind provider",
		RunE:  runKind,
		Args:  cobra.ExactArgs(1),
	}
	rk.Flags().UintP("nodes", "n", 1, "number of cluster nodes to start")
	rk.Flags().String("registry-port", "5000", "forwarded host port for registry container")
	rk.Flags().String("registry-proxy", "", "registry proxy to use")
	rk.Flags().String("ip-family", "", "ip family")
	rk.Flags().Bool("enable-cpu-manager", false, "enable cpu manager")
	return rk
}

func runKind(cmd *cobra.Command, args []string) error {
	nodes, err := cmd.Flags().GetUint("nodes")
	if err != nil {
		return err
	}
	port, err := cmd.Flags().GetString("registry-port")
	if err != nil {
		return err
	}
	rp, err := cmd.Flags().GetString("registry-proxy")
	if err != nil {
		return err
	}
	ipf, err := cmd.Flags().GetString("ip-family")
	if err != nil {
		return err
	}
	cpum, err := cmd.Flags().GetBool("enable-cpu-manager")
	if err != nil {
		return err
	}

	kindVersion := args[0]
	conf := &kind.KindConfig{
		Nodes:          int(nodes),
		Version:        kindVersion,
		RegistryPort:   port,
		RegistryProxy:  rp,
		WithCPUManager: cpum,
		IpFamily:       ipf,
	}

	switch kindVersion {
	case "k8s-1.28":
		kindProvider, err = kind.NewKindBaseProvider(conf)
		if err != nil {
			return err
		}
	case "sriov":
		kindProvider, err = sriov.NewKindSriovProvider(conf)
		if err != nil {
			return err
		}
	case "vgpu":
		kindProvider, err = vgpu.NewKindVGPU(conf)
	default:
		return fmt.Errorf("Invalid k8s version passed, please use one of k8s-1.28, sriov or vgpu")
	}

	b := context.Background()
	ctx, cancel := context.WithCancel(b)

	err = kindProvider.Start(ctx, cancel)
	if err != nil {
		return err
	}
	return nil
}
