package cmd

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/providers"
)

var nvmeDisks, scsiDisks, usbDisks, sharedDisks []string

func NewRunCommand() *cobra.Command {
	run := &cobra.Command{
		Use:   "run",
		Short: "run starts a given cluster",
		RunE:  run,
		Args:  cobra.ExactArgs(1),
	}
	run.Flags().UintP("nodes", "n", 1, "number of cluster nodes to start")
	run.Flags().UintP("numa", "u", 1, "number of NUMA nodes per node")
	run.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	run.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	run.Flags().UintP("secondary-nics", "", 0, "number of secondary nics to add")
	run.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	run.Flags().String("kernel-args", "", "additional kernel args to pass through to the nodes")
	run.Flags().BoolP("background", "b", true, "go to background after nodes are up")
	run.Flags().Bool("random-ports", true, "expose all ports on random localhost ports")
	run.Flags().Bool("slim", false, "use the slim flavor")
	run.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	run.Flags().Uint("http-port", 0, "port on localhost for http")
	run.Flags().Uint("https-port", 0, "port on localhost for https")
	run.Flags().Uint("registry-port", 0, "port on localhost for the docker registry")
	run.Flags().Uint("ocp-port", 0, "port on localhost for the ocp cluster")
	run.Flags().Uint("k8s-port", 0, "port on localhost for the k8s cluster")
	run.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")
	run.Flags().Uint("prometheus-port", 0, "port on localhost for prometheus server")
	run.Flags().Uint("grafana-port", 0, "port on localhost for grafana server")
	run.Flags().Uint("dns-port", 0, "port on localhost for dns server")
	run.Flags().String("nfs-data", "", "path to data which should be exposed via nfs to the nodes")
	run.Flags().Bool("enable-ceph", false, "enables dynamic storage provisioning using Ceph")
	run.Flags().Bool("enable-istio", false, "deploys Istio service mesh")
	run.Flags().Bool("enable-cnao", false, "enable network extensions with istio")
	run.Flags().Bool("skip-cnao-cr", false, "skip deploying cnao custom resource. if true, only cnao CRDS will be deployed")
	run.Flags().Bool("deploy-multus", false, "deploy multus")
	run.Flags().Bool("deploy-cdi", false, "deploy cdi")
	run.Flags().String("cdi-version", "", "cdi version")
	run.Flags().String("aaq-version", "", "aaq version")
	run.Flags().Bool("deploy-aaq", false, "deploy aaq")
	run.Flags().Bool("enable-nfs-csi", false, "deploys nfs csi dynamic storage")
	run.Flags().Bool("enable-prometheus", false, "deploys Prometheus operator")
	run.Flags().Bool("enable-prometheus-alertmanager", false, "deploys Prometheus alertmanager")
	run.Flags().Bool("enable-grafana", false, "deploys Grafana")
	run.Flags().Bool("enable-ksm", false, "enables kernel memory same page merging")
	run.Flags().Uint("ksm-page-count", 10, "number of pages to scan per time in ksm")
	run.Flags().Uint("ksm-scan-interval", 20, "sleep interval in milliseconds for ksm")
	run.Flags().Bool("enable-swap", false, "enable swap")
	run.Flags().Bool("unlimited-swap", false, "unlimited swap")
	run.Flags().Uint("swap-size", 0, "swap memory size in GB")
	run.Flags().Uint("swapiness", 0, "swapiness")
	run.Flags().String("docker-proxy", "", "sets network proxy for docker daemon")
	run.Flags().String("prefix", "", "container prefix")
	run.Flags().String("container-registry", "quay.io", "the registry to pull cluster container from")
	run.Flags().String("container-org", "kubevirtci", "the organization at the registry to pull the container from")
	run.Flags().String("container-suffix", "", "Override container suffix stored at the cli binary")
	run.Flags().String("gpu", "", "pci address of a GPU to assign to a node")
	run.Flags().StringArrayVar(&nvmeDisks, "nvme", []string{}, "size of the emulate NVMe disk to pass to the node")
	run.Flags().StringArrayVar(&scsiDisks, "scsi", []string{}, "size of the emulate SCSI disk to pass to the node")
	run.Flags().Bool("run-etcd-on-memory", false, "configure etcd to run on RAM memory, etcd data will not be persistent")
	run.Flags().String("etcd-capacity", "512M", "set etcd data mount size.\nthis flag takes affect only when 'run-etcd-on-memory' is specified")
	run.Flags().Uint("hugepages-2m", 64, "number of hugepages of size 2M to allocate")
	run.Flags().Uint("hugepages-1g", 0, "number of hugepages of size 1Gi to allocate")
	run.Flags().Bool("enable-realtime-scheduler", false, "configures the kernel to allow unlimited runtime for processes that require realtime scheduling")
	run.Flags().Bool("enable-fips", false, "enables FIPS")
	run.Flags().Bool("enable-psa", false, "Pod Security Admission")
	run.Flags().Bool("single-stack", false, "enable single stack IPv6")
	run.Flags().Bool("no-etcd-fsync", false, "unsafe: disable fsyncs in etcd")
	run.Flags().Bool("enable-audit", false, "enable k8s audit for all metadata events")
	run.Flags().StringArrayVar(&usbDisks, "usb", []string{}, "size of the emulate USB disk to pass to the node")
	run.Flags().StringArrayVar(&sharedDisks, "shared-block-device", []string{}, "size of block device to share between all nodes")

	return run
}

func run(cmd *cobra.Command, args []string) (retErr error) {
	opts := []providers.KubevirtProviderOption{}
	flags := cmd.Flags()
	for flagName, flagConfig := range providers.RunFlagMap {
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

	portMap := nat.PortMap{}
	utils.AppendTCPIfExplicit(portMap, utils.PortSSH, cmd.Flags(), "ssh-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortVNC, cmd.Flags(), "vnc-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortHTTP, cmd.Flags(), "http-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortHTTPS, cmd.Flags(), "https-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortAPI, cmd.Flags(), "k8s-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortOCP, cmd.Flags(), "ocp-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortRegistry, cmd.Flags(), "registry-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortPrometheus, cmd.Flags(), "prometheus-port")
	utils.AppendTCPIfExplicit(portMap, utils.PortGrafana, cmd.Flags(), "grafana-port")
	utils.AppendUDPIfExplicit(portMap, utils.PortDNS, cmd.Flags(), "dns-port")

	k8sVersion := args[0]
	allowedK8sVersions := []string{"k8s-1.28", "k8s-1.29", "k8s-1.30", "k8s-1.31"}
	var validVersion bool
	for _, v := range allowedK8sVersions {
		if k8sVersion == v {
			validVersion = true
		}
	}
	if !validVersion {
		return fmt.Errorf("Invalid k8s version passed, please use one of k8s-1.28, k8s-1.29, k8s-1.30 or k8s-1.31")
	}

	containerRegistry, err := cmd.Flags().GetString("container-registry")
	if err != nil {
		return err
	}

	containerOrg, err := cmd.Flags().GetString("container-org")
	if err != nil {
		return err
	}

	containerSuffix, err := cmd.Flags().GetString("container-suffix")
	if err != nil {
		return err
	}

	slim, err := cmd.Flags().GetBool("slim")
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	clusterImage := fmt.Sprintf("%s/%s/%s%s", containerRegistry, containerOrg, k8sVersion, containerSuffix)

	if slim {
		clusterImage += "-slim"
	}

	b := context.Background()
	ctx, cancel := context.WithCancel(b)
	err = docker.ImagePull(cli, ctx, clusterImage, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("Failed to download cluster image %s, %s", clusterImage, err)
	}
	kp := providers.NewKubevirtProvider(k8sVersion, clusterImage, cli, opts)

	err = kp.Start(ctx, cancel, portMap)
	if err != nil {
		return err
	}

	return nil
}
