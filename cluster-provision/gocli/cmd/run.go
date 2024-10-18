package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/resource"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/nodesconfig"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	containers2 "kubevirt.io/kubevirtci/cluster-provision/gocli/containers"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/images"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/aaq"
	bindvfio "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/bind-vfio"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/cdi"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/cnao"
	dockerproxy "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/docker-proxy"
	etcdinmemory "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/etcd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/istio"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/ksm"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/multus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/nfscsi"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/node01"
	nodesprovision "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/nodes"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/prometheus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/psa"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/realtime"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/rookceph"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/rootkey"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/swap"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"

	"github.com/alessio/shellescape"
)

const (
	proxySettings = `
curl {{.Proxy}}/ca.crt > /etc/pki/ca-trust/source/anchors/docker_registry_proxy.crt
update-ca-trust

mkdir -p /etc/systemd/system/crio.service.d
cat <<EOT >/etc/systemd/system/crio.service.d/override.conf
[Service]
Environment="HTTP_PROXY={{.Proxy}}"
Environment="HTTPS_PROXY={{.Proxy}}"
Environment="NO_PROXY=localhost,127.0.0.1,registry,10.96.0.0/12,10.244.0.0/16,192.168.0.0/16,fd00:10:96::/112,fd00:10:244::/112,fd00::/64"
EOT

systemctl daemon-reload
systemctl restart crio.service

while [[ systemctl status crio | grep active | wc -l -eq 0 ]]
do
    sleep 2
done
EOF
`
	etcdDataDir         = "/var/lib/etcd"
	nvmeDiskImagePrefix = "/nvme"
	scsiDiskImagePrefix = "/scsi"
)

var soundcardPCIIDs = []string{"8086:2668", "8086:2415"}
var cli *client.Client
var nvmeDisks []string
var scsiDisks []string
var usbDisks []string
var sshClient libssh.Client

type dockerSetting struct {
	Proxy string
}

// NewRunCommand returns command that runs given cluster
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
	run.Flags().Bool("reverse", false, "reverse node setup order")
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
	return run
}

func run(cmd *cobra.Command, args []string) (retErr error) {

	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}

	nodes, err := cmd.Flags().GetUint("nodes")
	if err != nil {
		return err
	}

	memory, err := cmd.Flags().GetString("memory")
	if err != nil {
		return err
	}
	resource.MustParse(memory)

	reverse, err := cmd.Flags().GetBool("reverse")
	if err != nil {
		return err
	}

	randomPorts, err := cmd.Flags().GetBool("random-ports")
	if err != nil {
		return err
	}

	slim, err := cmd.Flags().GetBool("slim")
	if err != nil {
		return err
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

	qemuArgs, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}
	kernelArgs, err := cmd.Flags().GetString("kernel-args")
	if err != nil {
		return err
	}

	cpu, err := cmd.Flags().GetUint("cpu")
	if err != nil {
		return err
	}

	numa, err := cmd.Flags().GetUint("numa")
	if err != nil {
		return err
	}

	secondaryNics, err := cmd.Flags().GetUint("secondary-nics")
	if err != nil {
		return err
	}

	nfsData, err := cmd.Flags().GetString("nfs-data")
	if err != nil {
		return err
	}

	dockerProxy, err := cmd.Flags().GetString("docker-proxy")
	if err != nil {
		return err
	}

	cephEnabled, err := cmd.Flags().GetBool("enable-ceph")
	if err != nil {
		return err
	}

	nfsCsiEnabled, err := cmd.Flags().GetBool("enable-nfs-csi")
	if err != nil {
		return err
	}

	istioEnabled, err := cmd.Flags().GetBool("enable-istio")
	if err != nil {
		return err
	}

	prometheusEnabled, err := cmd.Flags().GetBool("enable-prometheus")
	if err != nil {
		return err
	}

	prometheusAlertmanagerEnabled, err := cmd.Flags().GetBool("enable-prometheus-alertmanager")
	if err != nil {
		return err
	}

	grafanaEnabled, err := cmd.Flags().GetBool("enable-grafana")
	if err != nil {
		return err
	}

	cluster := args[0]

	background, err := cmd.Flags().GetBool("background")
	if err != nil {
		return err
	}

	containerRegistry, err := cmd.Flags().GetString("container-registry")
	if err != nil {
		return err
	}
	gpuAddress, err := cmd.Flags().GetString("gpu")
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

	runEtcdOnMemory, err := cmd.Flags().GetBool("run-etcd-on-memory")
	if err != nil {
		return err
	}

	etcdDataMountSize, err := cmd.Flags().GetString("etcd-capacity")
	if err != nil {
		return err
	}
	resource.MustParse(etcdDataMountSize)

	hugepages2Mcount, err := cmd.Flags().GetUint("hugepages-2m")
	if err != nil {
		return err
	}
	hugepages1Gcount, err := cmd.Flags().GetUint("hugepages-1g")
	if err != nil {
		return err
	}

	realtimeSchedulingEnabled, err := cmd.Flags().GetBool("enable-realtime-scheduler")
	if err != nil {
		return err
	}
	psaEnabled, err := cmd.Flags().GetBool("enable-psa")
	if err != nil {
		return err
	}
	singleStack, err := cmd.Flags().GetBool("single-stack")
	if err != nil {
		return err
	}
	noEtcdFsync, err := cmd.Flags().GetBool("no-etcd-fsync")
	if err != nil {
		return err
	}
	enableAudit, err := cmd.Flags().GetBool("enable-audit")
	if err != nil {
		return err
	}
	fipsEnabled, err := cmd.Flags().GetBool("enable-fips")
	if err != nil {
		return err
	}
	cnaoEnabled, err := cmd.Flags().GetBool("enable-cnao")
	if err != nil {
		return err
	}
	cnaoSkipCR, err := cmd.Flags().GetBool("skip-cnao-cr")
	if err != nil {
		return err
	}

	deployCdi, err := cmd.Flags().GetBool("deploy-cdi")
	if err != nil {
		return err
	}

	cdiVersion, err := cmd.Flags().GetString("cdi-version")
	if err != nil {
		return err
	}

	deployAaq, err := cmd.Flags().GetBool("deploy-aaq")
	if err != nil {
		return err
	}

	aaqVersion, err := cmd.Flags().GetString("aaq-version")
	if err != nil {
		return err
	}

	deployMultus, err := cmd.Flags().GetBool("deploy-multus")
	if err != nil {
		return err
	}

	enableSwap, err := cmd.Flags().GetBool("enable-swap")
	if err != nil {
		return err
	}

	unlimitedSwap, err := cmd.Flags().GetBool("unlimited-swap")
	if err != nil {
		return err
	}

	swapiness, err := cmd.Flags().GetUint("swapiness")
	if err != nil {
		return err
	}

	swapSize, err := cmd.Flags().GetUint("swap-size")
	if err != nil {
		return err
	}

	enableKsm, err := cmd.Flags().GetBool("enable-ksm")
	if err != nil {
		return err
	}

	ksmPageCount, err := cmd.Flags().GetUint("ksm-page-count")
	if err != nil {
		return err
	}

	ksmScanInterval, err := cmd.Flags().GetUint("ksm-scan-interval")
	if err != nil {
		return err
	}

	cli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	b := context.Background()
	ctx, cancel := context.WithCancel(b)

	stop := make(chan error, 10)
	containers, _, done := docker.NewCleanupHandler(cli, stop, cmd.OutOrStderr(), false)

	defer func() {
		stop <- retErr
		<-done
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
		cancel()
		stop <- fmt.Errorf("Interrupt received, clean up")
	}()

	clusterImage := cluster

	// Check if cluster container suffix has not being override
	// in that case use the default prefix stored at the binary
	if containerSuffix == "" {
		containerSuffix = images.SUFFIX
	}
	if containerSuffix != "" {
		clusterImage = fmt.Sprintf("%s/%s%s", containerOrg, cluster, containerSuffix)
	} else {
		clusterImage = path.Join(containerOrg, cluster)
	}

	if slim {
		clusterImage += "-slim"
	}

	if len(containerRegistry) > 0 {
		clusterImage = path.Join(containerRegistry, clusterImage)
		fmt.Printf("Download the image %s\n", clusterImage)
		err = docker.ImagePull(cli, ctx, clusterImage, types.ImagePullOptions{})
		if err != nil {
			panic(fmt.Sprintf("Failed to download cluster image %s, %s", clusterImage, err))
		}
	}

	var dnsmasq *container.CreateResponse
	for i := 0; i <= 3; i++ {
		if i == 3 {
			fmt.Printf("dnsmasq container failed to start 3 times")
			return err
		}
		dnsmasq, err = containers2.DNSMasq(cli, ctx, &containers2.DNSMasqOptions{
			ClusterImage:       clusterImage,
			SecondaryNicsCount: secondaryNics,
			RandomPorts:        randomPorts,
			PortMap:            portMap,
			Prefix:             prefix,
			NodeCount:          nodes,
		})
		if err != nil {
			return err
		}

		if err := cli.ContainerStart(ctx, dnsmasq.ID, container.StartOptions{}); err != nil {
			fmt.Printf("Failed to start dnsmasq container: %s\n", err)
			fmt.Printf("Retry creating and starting dnsmasq container\n")
			if err := cli.ContainerRemove(ctx, dnsmasq.ID, container.RemoveOptions{}); err != nil {
				return err
			}
			time.Sleep(2 * time.Second)

		} else {
			containers <- dnsmasq.ID
			break
		}
	}

	dm, err := cli.ContainerInspect(context.Background(), dnsmasq.ID)
	if err != nil {
		return err
	}

	sshPort, err := utils.GetPublicPort(utils.PortSSH, dm.NetworkSettings.Ports)
	apiServerPort, err := utils.GetPublicPort(utils.PortAPI, dm.NetworkSettings.Ports)

	// Pull the registry image
	err = docker.ImagePull(cli, ctx, utils.DockerRegistryImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	// Start registry
	registry, err := cli.ContainerCreate(ctx, &container.Config{
		Image: utils.DockerRegistryImage,
	}, &container.HostConfig{
		Privileged:  true, // fixme we just need proper selinux volume labeling
		NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
	}, nil, nil, prefix+"-registry")
	if err != nil {
		return err
	}
	containers <- registry.ID
	if err := cli.ContainerStart(ctx, registry.ID, container.StartOptions{}); err != nil {
		return err
	}

	if nfsData != "" {
		nfsData, err := filepath.Abs(nfsData)
		if err != nil {
			return err
		}
		// Pull the ganesha image
		err = docker.ImagePull(cli, ctx, utils.NFSGaneshaImage, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		// Start the ganesha image
		nfsServer, err := cli.ContainerCreate(ctx, &container.Config{
			Image: utils.NFSGaneshaImage,
		}, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: nfsData,
					Target: "/data/nfs",
				},
			},
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
		}, nil, nil, prefix+"-nfs-ganesha")
		if err != nil {
			return err
		}
		containers <- nfsServer.ID
		if err := cli.ContainerStart(ctx, nfsServer.ID, container.StartOptions{}); err != nil {
			return err
		}
	}

	// Add serial pty so we can do stuff like 'screen /dev/pts0' to access
	// the VM console from the container without ssh
	qemuArgs += " -serial pty"

	wg := sync.WaitGroup{}
	wg.Add(int(nodes))
	// start one vm after each other
	macCounter := 0
	for x := 0; x < int(nodes); x++ {

		nodeQemuArgs := qemuArgs

		for i := 0; i < int(secondaryNics); i++ {
			netSuffix := fmt.Sprintf("%d-%d", x, i)
			macSuffix := fmt.Sprintf("%02x", macCounter)
			macCounter++
			nodeQemuArgs = fmt.Sprintf("%s -device virtio-net-pci,netdev=secondarynet%s,mac=52:55:00:d1:56:%s -netdev tap,id=secondarynet%s,ifname=stap%s,script=no,downscript=no", nodeQemuArgs, netSuffix, macSuffix, netSuffix, netSuffix)
		}

		nodeName := nodeNameFromIndex(x + 1)
		nodeNum := fmt.Sprintf("%02d", x+1)
		sshClient, err = libssh.NewSSHClient(sshPort, x+1, false)
		if err != nil {
			return err
		}
		if reverse {
			nodeName = nodeNameFromIndex((int(nodes) - x))
			nodeNum = fmt.Sprintf("%02d", (int(nodes) - x))
			sshClient, err = libssh.NewSSHClient(sshPort, (int(nodes) - x), false)

			if err != nil {
				return err
			}
		}

		// assign a GPU to one node
		var deviceMappings []container.DeviceMapping
		if gpuAddress != "" && x == int(nodes)-1 {
			iommu_group, err := getPCIDeviceIOMMUGroup(gpuAddress)
			if err != nil {
				return err
			}
			vfioDevice := fmt.Sprintf("/dev/vfio/%s", iommu_group)
			deviceMappings = []container.DeviceMapping{
				{
					PathOnHost:        "/dev/vfio/vfio",
					PathInContainer:   "/dev/vfio/vfio",
					CgroupPermissions: "mrw",
				},
				{
					PathOnHost:        vfioDevice,
					PathInContainer:   vfioDevice,
					CgroupPermissions: "mrw",
				},
			}
			nodeQemuArgs = fmt.Sprintf("%s -device vfio-pci,host=%s", nodeQemuArgs, gpuAddress)
		}

		var vmArgsNvmeDisks []string
		if len(nvmeDisks) > 0 {
			for i, size := range nvmeDisks {
				resource.MustParse(size)
				disk := fmt.Sprintf("%s-%d.img", nvmeDiskImagePrefix, i)
				nodeQemuArgs = fmt.Sprintf("%s -drive file=%s,format=raw,id=NVME%d,if=none -device nvme,drive=NVME%d,serial=nvme-%d", nodeQemuArgs, disk, i, i, i)
				vmArgsNvmeDisks = append(vmArgsNvmeDisks, fmt.Sprintf("--nvme-device-size %s", size))
			}
		}
		var vmArgsSCSIDisks []string
		if len(scsiDisks) > 0 {
			nodeQemuArgs = fmt.Sprintf("%s -device virtio-scsi-pci,id=scsi0", nodeQemuArgs)
			for i, size := range scsiDisks {
				resource.MustParse(size)
				disk := fmt.Sprintf("%s-%d.img", scsiDiskImagePrefix, i)
				nodeQemuArgs = fmt.Sprintf("%s -drive file=%s,if=none,id=drive%d -device scsi-hd,drive=drive%d,bus=scsi0.0,channel=0,scsi-id=0,lun=%d", nodeQemuArgs, disk, i, i, i)
				vmArgsSCSIDisks = append(vmArgsSCSIDisks, fmt.Sprintf("--scsi-device-size %s", size))
			}
		}

		var vmArgsUSBDisks []string
		const bus = " -device qemu-xhci,id=bus%d"
		const drive = " -drive if=none,id=stick%d,format=raw,file=/usb-%d.img"
		const dev = " -device usb-storage,bus=bus%d.0,drive=stick%d"
		const usbSizefmt = " --usb-device-size %s"
		if len(usbDisks) > 0 {
			for i, size := range usbDisks {
				resource.MustParse(size)
				if i%2 == 0 {
					nodeQemuArgs += fmt.Sprintf(bus, i/2)
				}
				nodeQemuArgs += fmt.Sprintf(drive, i, i)
				nodeQemuArgs += fmt.Sprintf(dev, i/2, i)
				vmArgsUSBDisks = append(vmArgsUSBDisks, fmt.Sprintf(usbSizefmt, size))
			}
		}

		additionalArgs := []string{}
		if len(nodeQemuArgs) > 0 {
			additionalArgs = append(additionalArgs, "--qemu-args", shellescape.Quote(nodeQemuArgs))
		}

		if hugepages2Mcount > 0 {
			kernelArgs += fmt.Sprintf(" hugepagesz=2M hugepages=%d", hugepages2Mcount)
		}

		if hugepages1Gcount > 0 {
			kernelArgs += fmt.Sprintf(" hugepagesz=1G hugepages=%d", hugepages1Gcount)
		}

		if fipsEnabled {
			kernelArgs += " fips=1"
		}

		blockDev := ""
		if cephEnabled {
			blockDev = "--block-device /var/run/disk/blockdev.qcow2 --block-device-size 32212254720"
		}

		kernelArgs = strings.TrimSpace(kernelArgs)
		if kernelArgs != "" {
			additionalArgs = append(additionalArgs, "--additional-kernel-args", shellescape.Quote(kernelArgs))
		}

		vmContainerConfig := &container.Config{
			Image: clusterImage,
			Env: []string{
				fmt.Sprintf("NODE_NUM=%s", nodeNum),
			},
			Cmd: []string{"/bin/bash", "-c", fmt.Sprintf("/vm.sh -n /var/run/disk/disk.qcow2 --memory %s --cpu %s --numa %s %s %s %s %s %s",
				memory,
				strconv.Itoa(int(cpu)),
				strconv.Itoa(int(numa)),
				blockDev,
				strings.Join(vmArgsSCSIDisks, " "),
				strings.Join(vmArgsNvmeDisks, " "),
				strings.Join(vmArgsUSBDisks, " "),
				strings.Join(additionalArgs, " "),
			)},
		}

		if cephEnabled {
			vmContainerConfig.Volumes = map[string]struct{}{
				"/var/lib/rook": {},
			}
		}

		node, err := cli.ContainerCreate(ctx, vmContainerConfig, &container.HostConfig{
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
			Resources: container.Resources{
				Devices: deviceMappings,
			},
		}, nil, nil, prefix+"-"+nodeName)
		if err != nil {
			return err
		}
		containers <- node.ID
		if err := cli.ContainerStart(ctx, node.ID, container.StartOptions{}); err != nil {
			return err
		}

		// Wait for vm start
		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "while [ ! -f /ssh_ready ] ; do sleep 1; done"}, os.Stdout)
		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("checking for ssh.sh script for node %s failed", nodeName)
		}

		err = waitForVMToBeUp(prefix, nodeName)
		if err != nil {
			return err
		}

		rootkey := rootkey.NewRootKey(sshClient)
		if err = rootkey.Exec(); err != nil {
			return err
		}
		sshClient, err = libssh.NewSSHClient(sshPort, x+1, true)

		linuxConfigFuncs := []nodesconfig.LinuxConfigFunc{
			nodesconfig.WithFipsEnabled(fipsEnabled),
			nodesconfig.WithDockerProxy(dockerProxy),
			nodesconfig.WithEtcdInMemory(runEtcdOnMemory),
			nodesconfig.WithEtcdSize(etcdDataMountSize),
			nodesconfig.WithSingleStack(singleStack),
			nodesconfig.WithNoEtcdFsync(noEtcdFsync),
			nodesconfig.WithEnableAudit(enableAudit),
			nodesconfig.WithGpuAddress(gpuAddress),
			nodesconfig.WithRealtime(realtimeSchedulingEnabled),
			nodesconfig.WithPSA(psaEnabled),
			nodesconfig.WithKsm(enableKsm),
			nodesconfig.WithKsmPageCount(int(ksmPageCount)),
			nodesconfig.WithKsmScanInterval(int(ksmScanInterval)),
			nodesconfig.WithSwap(enableSwap),
			nodesconfig.WithSwapiness(int(swapiness)),
			nodesconfig.WithSwapSize(int(swapSize)),
			nodesconfig.WithUnlimitedSwap(unlimitedSwap),
		}

		n := nodesconfig.NewNodeLinuxConfig(x+1, prefix, linuxConfigFuncs)

		if err = provisionNode(sshClient, n); err != nil {
			return err
		}

		go func(id string) {
			cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
			wg.Done()
		}(node.ID)
	}

	sshClient, err := libssh.NewSSHClient(sshPort, 1, true)
	if err != nil {
		return err
	}

	k8sConfs := []nodesconfig.K8sConfigFunc{
		nodesconfig.WithCeph(cephEnabled),
		nodesconfig.WithPrometheus(prometheusEnabled),
		nodesconfig.WithAlertmanager(prometheusAlertmanagerEnabled),
		nodesconfig.WithGrafana(grafanaEnabled),
		nodesconfig.WithIstio(istioEnabled),
		nodesconfig.WithNfsCsi(nfsCsiEnabled),
		nodesconfig.WithCnao(cnaoEnabled),
		nodesconfig.WithCNAOSkipCR(cnaoSkipCR),
		nodesconfig.WithMultus(deployMultus),
		nodesconfig.WithCdi(deployCdi),
		nodesconfig.WithCdiVersion(cdiVersion),
		nodesconfig.WithAAQ(deployAaq),
		nodesconfig.WithAAQVersion(aaqVersion),
	}
	n := nodesconfig.NewNodeK8sConfig(k8sConfs)

	kubeConfFile, err := os.Create(".kubeconfig")
	if err != nil {
		return err
	}

	err = sshClient.CopyRemoteFile("/etc/kubernetes/admin.conf", kubeConfFile)
	if err != nil {
		return err
	}

	config, err := k8s.NewConfig(".kubeconfig", apiServerPort)
	if err != nil {
		return err
	}

	k8sClient, err := k8s.NewDynamicClient(config)
	if err != nil {
		return err
	}

	if err = provisionK8sOptions(sshClient, k8sClient, n, prefix); err != nil {
		return err
	}

	// If background flag was specified, we don't want to clean up if we reach that state
	if !background {
		wg.Wait()
		stop <- fmt.Errorf("Done. please clean up")
	}

	return nil
}

func provisionK8sOptions(sshClient libssh.Client, k8sClient k8s.K8sDynamicClient, n *nodesconfig.NodeK8sConfig, k8sVersion string) error {
	opts := []opts.Opt{}

	if n.Ceph {
		cephOpt := rookceph.NewCephOpt(k8sClient, sshClient)
		opts = append(opts, cephOpt)
	}

	if n.NfsCsi {
		nfsCsiOpt := nfscsi.NewNfsCsiOpt(k8sClient)
		opts = append(opts, nfsCsiOpt)
	}

	if n.Multus {
		multusOpt := multus.NewMultusOpt(k8sClient, sshClient)
		opts = append(opts, multusOpt)
	}

	if n.CNAO {
		cnaoOpt := cnao.NewCnaoOpt(k8sClient, sshClient, n.Multus, n.CNAOSkipCR)
		opts = append(opts, cnaoOpt)
	}

	if n.Istio {
		istioOpt := istio.NewIstioOpt(sshClient, k8sClient, n.CNAO)
		opts = append(opts, istioOpt)
	}

	if n.Prometheus {
		prometheusOpt := prometheus.NewPrometheusOpt(k8sClient, n.Grafana, n.Alertmanager)
		opts = append(opts, prometheusOpt)
	}

	if n.CDI {
		cdi := cdi.NewCdiOpt(k8sClient, sshClient, n.CDIVersion)
		opts = append(opts, cdi)
	}

	if n.AAQ {
		if k8sVersion == "k8s-1.30" {
			aaq := aaq.NewAaqOpt(k8sClient, sshClient, n.CDIVersion)
			opts = append(opts, aaq)
		} else {
			logrus.Info("AAQ was requested but k8s version is not k8s-1.30, skipping")
		}
	}

	for _, opt := range opts {
		if err := opt.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func provisionNode(sshClient libssh.Client, n *nodesconfig.NodeLinuxConfig) error {
	opts := []opts.Opt{}
	nodeName := nodeNameFromIndex(n.NodeIdx)

	if n.FipsEnabled {
		for _, cmd := range []string{"sudo fips-mode-setup --enable", "sudo reboot"} {
			if err := sshClient.Command(cmd); err != nil {
				return fmt.Errorf("Starting fips mode failed: %s", err)
			}
		}
		err := waitForVMToBeUp(n.K8sVersion, nodeName)
		if err != nil {
			return err
		}
	}

	if n.DockerProxy != "" {
		//if dockerProxy has value, generate a shell script`/script/docker-proxy.sh` which can be applied to set proxy settings
		dp := dockerproxy.NewDockerProxyOpt(sshClient, n.DockerProxy)
		opts = append(opts, dp)
	}

	if n.EtcdInMemory {
		logrus.Infof("Creating in-memory mount for etcd data on node %s", nodeName)
		etcdinmem := etcdinmemory.NewEtcdInMemOpt(sshClient, n.EtcdSize)
		opts = append(opts, etcdinmem)
	}

	if n.Realtime {
		realtimeOpt := realtime.NewRealtimeOpt(sshClient)
		opts = append(opts, realtimeOpt)
	}

	for _, s := range soundcardPCIIDs {
		// move the VM sound cards to a vfio-pci driver to prepare for assignment
		bvfio := bindvfio.NewBindVfioOpt(sshClient, s)
		opts = append(opts, bvfio)
	}

	if n.EnableAudit {
		if err := sshClient.Command("touch /home/vagrant/enable_audit"); err != nil {
			return fmt.Errorf("provisioning node %d failed (setting enableAudit phase): %s", n.NodeIdx, err)
		}
	}

	if n.PSA {
		psaOpt := psa.NewPsaOpt(sshClient)
		opts = append(opts, psaOpt)
	}

	if n.NodeIdx == 1 {
		n := node01.NewNode01Provisioner(sshClient, n.SingleStack, n.NoEtcdFsync)
		opts = append(opts, n)

	} else {
		if n.GpuAddress != "" {
			// move the assigned PCI device to a vfio-pci driver to prepare for assignment
			gpuDeviceID, err := getDevicePCIID(n.GpuAddress)
			if err != nil {
				return err
			}
			bindVfioOpt := bindvfio.NewBindVfioOpt(sshClient, gpuDeviceID)
			opts = append(opts, bindVfioOpt)
		}
		n := nodesprovision.NewNodesProvisioner(sshClient, n.SingleStack)
		opts = append(opts, n)
	}

	if n.KsmEnabled {
		ksmOpt := ksm.NewKsmOpt(sshClient, n.KsmScanInterval, n.KsmPageCount)
		opts = append(opts, ksmOpt)
	}

	if n.SwapEnabled {
		swapOpt := swap.NewSwapOpt(sshClient, n.Swappiness, n.UnlimitedSwap, n.SwapSize)
		opts = append(opts, swapOpt)
	}

	for _, o := range opts {
		if err := o.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func waitForVMToBeUp(prefix string, nodeName string) error {
	var err error
	// Wait for the VM to be up
	for x := 0; x < 10; x++ {
		err = _cmd(cli, nodeContainer(prefix, nodeName), "ssh.sh echo VM is up", "waiting for node to come up")
		if err == nil {
			break
		}
		logrus.WithError(err).Warningf("Could not establish a ssh connection to the VM, retrying ...")
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("could not establish a connection to the node after a generous timeout: %v", err)
	}

	return nil
}

func nodeNameFromIndex(x int) string {
	return fmt.Sprintf("node%02d", x)
}

func nodeContainer(prefix string, node string) string {
	return prefix + "-" + node
}

func getDockerProxyConfig(proxy string) (string, error) {
	p := dockerSetting{Proxy: proxy}
	buf := new(bytes.Buffer)

	t, err := template.New("docker-proxy").Parse(proxySettings)
	if err != nil {
		return "", err
	}
	err = t.Execute(buf, p)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// getDeviceIOMMUGroup gets devices iommu_group
// e.g. /sys/bus/pci/devices/0000\:65\:00.0/iommu_group -> ../../../../../kernel/iommu_groups/45
func getPCIDeviceIOMMUGroup(pciAddress string) (string, error) {
	iommuLink := filepath.Join("/sys/bus/pci/devices", pciAddress, "iommu_group")
	iommuPath, err := os.Readlink(iommuLink)
	if err != nil {
		return "", fmt.Errorf("failed to read iommu_group link %s for device %s - %v", iommuLink, pciAddress, err)
	}
	_, iommuGroup := filepath.Split(iommuPath)
	return iommuGroup, nil
}

func getDevicePCIID(pciAddress string) (string, error) {
	file, err := os.Open(filepath.Join("/sys/bus/pci/devices", pciAddress, "uevent"))
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PCI_ID") {
			equal := strings.Index(line, "=")
			value := strings.TrimSpace(line[equal+1:])
			return strings.ToLower(value), nil
		}
	}
	return "", fmt.Errorf("no pci_id is found")
}
