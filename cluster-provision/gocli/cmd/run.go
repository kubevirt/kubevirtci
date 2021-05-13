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
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/resource"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	containers2 "kubevirt.io/kubevirtci/cluster-provision/gocli/containers"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/images"
)

const soundcardPCIID = "8086:2668"
const proxySettings = `
mkdir -p /etc/systemd/system/docker.service.d/

cat <<EOT  >/etc/systemd/system/docker.service.d/proxy.conf
[Service]
Environment="HTTP_PROXY={{.Proxy}}"
Environment="HTTPS_PROXY={{.Proxy}}"
Environment="NO_PROXY=localhost,127.0.0.1"
EOT

systemctl daemon-reload
systemctl restart docker
EOF
`
const etcdDataDir = "/var/lib/etcd"

var cli *client.Client

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
	run.Flags().StringP("memory", "m", "3096M", "amount of ram per node")
	run.Flags().UintP("cpu", "c", 2, "number of cpu cores per node")
	run.Flags().UintP("secondary-nics", "", 0, "number of secondary nics to add")
	run.Flags().String("qemu-args", "", "additional qemu args to pass through to the nodes")
	run.Flags().BoolP("background", "b", false, "go to background after nodes are up")
	run.Flags().BoolP("reverse", "r", false, "revert node startup order")
	run.Flags().Bool("random-ports", true, "expose all ports on random localhost ports")
	run.Flags().String("registry-volume", "", "cache docker registry content in the specified volume")
	run.Flags().Uint("vnc-port", 0, "port on localhost for vnc")
	run.Flags().Uint("http-port", 0, "port on localhost for http")
	run.Flags().Uint("https-port", 0, "port on localhost for https")
	run.Flags().Uint("registry-port", 0, "port on localhost for the docker registry")
	run.Flags().Uint("ocp-port", 0, "port on localhost for the ocp cluster")
	run.Flags().Uint("k8s-port", 0, "port on localhost for the k8s cluster")
	run.Flags().Uint("ssh-port", 0, "port on localhost for ssh server")
	run.Flags().String("nfs-data", "", "path to data which should be exposed via nfs to the nodes")
	run.Flags().String("log-to-dir", "", "enables aggregated cluster logging to the folder")
	run.Flags().Bool("enable-ceph", false, "enables dynamic storage provisioning using Ceph")
	run.Flags().Bool("enable-istio", false, "deploys Istio service mesh")
	run.Flags().Bool("enable-prometheus", false, "deploys Prometheus operator")
	run.Flags().UintP("prometheus-replicas", "x", 1, "number of prometheus replicas to start")
	run.Flags().Bool("enable-prometheus-alertmanager", false, "deploys Prometheus alertmanager")
	run.Flags().UintP("prometheus-alertmanager-replicas", "y", 1, "number of prometheus alertmanager replicas to start")
	run.Flags().Bool("enable-grafana", false, "deploys Grafana")
	run.Flags().String("docker-proxy", "", "sets network proxy for docker daemon")
	run.Flags().String("container-registry", "quay.io", "the registry to pull cluster container from")
	run.Flags().String("container-org", "kubevirtci", "the organization at the registry to pull the container from")
	run.Flags().String("container-suffix", "", "Override container suffix stored at the cli binary")
	run.Flags().String("gpu", "", "pci address of a GPU to assign to a node")
	run.Flags().Bool("run-etcd-on-memory", false, "configure etcd to run on RAM memory, etcd data will not be persistent")
	run.Flags().String("etcd-capacity", "512M", "set etcd data mount size.\nthis flag takes affect only when 'run-etcd-on-memory' is specified")

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

	portMap := nat.PortMap{}

	utils.AppendIfExplicit(portMap, utils.PortSSH, cmd.Flags(), "ssh-port")
	utils.AppendIfExplicit(portMap, utils.PortVNC, cmd.Flags(), "vnc-port")
	utils.AppendIfExplicit(portMap, utils.PortHTTP, cmd.Flags(), "http-port")
	utils.AppendIfExplicit(portMap, utils.PortHTTPS, cmd.Flags(), "https-port")
	utils.AppendIfExplicit(portMap, utils.PortAPI, cmd.Flags(), "k8s-port")
	utils.AppendIfExplicit(portMap, utils.PortOCP, cmd.Flags(), "ocp-port")
	utils.AppendIfExplicit(portMap, utils.PortRegistry, cmd.Flags(), "registry-port")

	qemuArgs, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}

	cpu, err := cmd.Flags().GetUint("cpu")
	if err != nil {
		return err
	}

	secondaryNics, err := cmd.Flags().GetUint("secondary-nics")
	if err != nil {
		return err
	}

	registryVol, err := cmd.Flags().GetString("registry-volume")
	if err != nil {
		return err
	}

	nfsData, err := cmd.Flags().GetString("nfs-data")
	if err != nil {
		return err
	}

	logDir, err := cmd.Flags().GetString("log-to-dir")
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

	istioEnabled, err := cmd.Flags().GetBool("enable-istio")
	if err != nil {
		return err
	}

	prometheusEnabled, err := cmd.Flags().GetBool("enable-prometheus")
	if err != nil {
		return err
	}

	prometheusReplicas, err := cmd.Flags().GetUint("prometheus-replicas")
	if err != nil {
		return err
	}

	prometheusAlertmanagerEnabled, err := cmd.Flags().GetBool("enable-prometheus-alertmanager")
	if err != nil {
		return err
	}

	prometheusAlertmanagerReplicas, err := cmd.Flags().GetUint("prometheus-alertmanager-replicas")
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

	cli, err = client.NewEnvClient()
	if err != nil {
		return err
	}

	b := context.Background()
	ctx, cancel := context.WithCancel(b)

	stop := make(chan error, 10)
	containers, volumes, done := docker.NewCleanupHandler(cli, stop, cmd.OutOrStderr(), false)

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
	// in that case use the default preffix stored at the binary
	if containerSuffix == "" {
		containerSuffix = images.SUFFIX
	}
	if containerSuffix != "" {
		clusterImage = fmt.Sprintf("%s/%s%s", containerOrg, cluster, containerSuffix)
	} else {
		clusterImage = path.Join(containerOrg, cluster)
	}

	if len(containerRegistry) > 0 {
		clusterImage = path.Join(containerRegistry, clusterImage)
		fmt.Printf("Download the image %s\n", clusterImage)
		err = docker.ImagePull(cli, ctx, clusterImage, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}
	}

	dnsmasq, err := containers2.DNSMasq(cli, ctx, &containers2.DNSMasqOptions{
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

	containers <- dnsmasq.ID
	if err := cli.ContainerStart(ctx, dnsmasq.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Pull the registry image
	err = docker.ImagePull(cli, ctx, utils.DockerRegistryImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	// Create registry volume
	var registryMounts []mount.Mount
	if registryVol != "" {

		vol, err := cli.VolumeCreate(ctx, volume.VolumeCreateBody{
			Name: fmt.Sprintf("%s-%s", prefix, "registry"),
		})
		if err != nil {
			return err
		}
		registryMounts = []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: vol.Name,
				Target: "/var/lib/registry",
			},
		}
	}

	// Start registry
	registry, err := cli.ContainerCreate(ctx, &container.Config{
		Image: utils.DockerRegistryImage,
	}, &container.HostConfig{
		Mounts:      registryMounts,
		Privileged:  true, // fixme we just need proper selinux volume labeling
		NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
	}, nil, nil, prefix+"-registry")
	if err != nil {
		return err
	}
	containers <- registry.ID
	if err := cli.ContainerStart(ctx, registry.ID, types.ContainerStartOptions{}); err != nil {
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
		if err := cli.ContainerStart(ctx, nfsServer.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
	}

	if logDir != "" {
		logDir, err := filepath.Abs(logDir)
		if err != nil {
			return err
		}

		if _, err = os.Stat(logDir); os.IsNotExist(err) {
			os.Mkdir(logDir, 0755)
		}

		// Pull the fluent image
		err = docker.ImagePull(cli, ctx, utils.FluentdImage, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		// Start the fluent image
		fluentd, err := cli.ContainerCreate(ctx, &container.Config{
			Image: utils.FluentdImage,
			Cmd: strslice.StrSlice{
				"exec fluentd",
				"-i \"<system>\n log_level debug\n</system>\n<source>\n@type  forward\n@log_level error\nport  24224\n</source>\n<match **>\n@type file\npath /fluentd/log/collected\n</match>\"",
				"-p /fluentd/plugins $FLUENTD_OPT -v",
			},
		}, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: logDir,
					Target: "/fluentd/log/collected",
				},
			},
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + dnsmasq.ID),
		}, nil, nil, prefix+"-fluentd")
		if err != nil {
			return err
		}
		containers <- fluentd.ID
		if err := cli.ContainerStart(ctx, fluentd.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}
	}

	// Add serial pty so we can do stuff like 'socat - /dev/pts0' to access
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
		if reverse {
			nodeName = nodeNameFromIndex((int(nodes) - x))
			nodeNum = fmt.Sprintf("%02d", (int(nodes) - x))
		}

		vol, err := cli.VolumeCreate(ctx, volume.VolumeCreateBody{
			Name: fmt.Sprintf("%s-%s", prefix, nodeName),
		})
		if err != nil {
			return err
		}
		volumes <- vol.Name

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

		if len(nodeQemuArgs) > 0 {
			nodeQemuArgs = "--qemu-args \"" + nodeQemuArgs + "\""
		}

		blockDev := ""
		if cephEnabled {
			blockDev = "--block-device /var/run/disk/blockdev.qcow2 --block-device-size 32212254720"
		}

		vmContainerConfig := &container.Config{
			Image: clusterImage,
			Env: []string{
				fmt.Sprintf("NODE_NUM=%s", nodeNum),
			},
			Cmd: []string{"/bin/bash", "-c", fmt.Sprintf("/vm.sh -n /var/run/disk/disk.qcow2 --memory %s --cpu %s %s %s", memory, strconv.Itoa(int(cpu)), blockDev, nodeQemuArgs)},
		}

		if cephEnabled {
			vmContainerConfig.Volumes = map[string]struct{}{
				"/var/lib/rook": struct{}{},
			}
		}

		node, err := cli.ContainerCreate(ctx, vmContainerConfig, &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   "volume",
					Source: vol.Name,
					Target: "/var/run/disk",
				},
			},
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
		if err := cli.ContainerStart(ctx, node.ID, types.ContainerStartOptions{}); err != nil {
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

		if dockerProxy != "" {
			//if dockerProxy has value, genterate a shell script`/script/docker-proxy.sh` which can be applied to set proxy settings
			proxyConfig, err := getDockerProxyConfig(dockerProxy)
			if err != nil {
				return fmt.Errorf("parsing proxy settings for node %s failed", nodeName)
			}
			success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("cat <<EOF >/scripts/docker-proxy.sh %s", proxyConfig)}, os.Stdout)
			if err != nil {
				return fmt.Errorf("write failed for proxy provision script for node %s", nodeName)
			}
			if success {
				success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo /bin/bash < /scripts/docker-proxy.sh")}, os.Stdout)
			}
		}

		if runEtcdOnMemory {
			logrus.Infof("Creating in-memory mount for etcd data on node %s", nodeName)
			err = prepareEtcdDataMount(nodeContainer(prefix, nodeName), etcdDataDir, etcdDataMountSize)
			if err != nil {
				logrus.Errorf("failed to create mount for etcd data on node %s: %v", nodeName, err)
				return err
			}
		}

		//check if we have a special provision script
		success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("test -f /scripts/%s.sh", nodeName)}, os.Stdout)
		if err != nil {
			return fmt.Errorf("checking for matching provision script for node %s failed", nodeName)
		}

		if success {
			// move the VM sound card to a vfio-pci driver to prepare for assignment
			err = prepareDeviceForAssignment(cli, nodeContainer(prefix, nodeName), soundcardPCIID, "")
			if err != nil {
				return err
			}
			success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo /bin/bash < /scripts/%s.sh", nodeName)}, os.Stdout)
		} else {
			// move the VM sound card to a vfio-pci driver to prepare for assignment
			err = prepareDeviceForAssignment(cli, nodeContainer(prefix, nodeName), soundcardPCIID, "")
			if err != nil {
				return err
			}
			if gpuAddress != "" {
				// move the assigned PCI device to a vfio-pci driver to prepare for assignment
				err = prepareDeviceForAssignment(cli, nodeContainer(prefix, nodeName), "", gpuAddress)
				if err != nil {
					return err
				}
			}
			success, err = docker.Exec(cli, nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "ssh.sh sudo /bin/bash < /scripts/nodes.sh"}, os.Stdout)
		}

		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("provisioning node %s failed", nodeName)
		}

		go func(id string) {
			cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
			wg.Done()
		}(node.ID)
	}

	if cephEnabled {
		nodeName := nodeNameFromIndex(1)
		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{
			"/bin/bash",
			"-c",
			"ssh.sh sudo /bin/bash < /scripts/rook-ceph.sh",
		}, os.Stdout)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("provisioning Ceph CSI failed")
		}
	}

	if istioEnabled {
		nodeName := nodeNameFromIndex(1)
		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{
			"/bin/bash",
			"-c",
			"ssh.sh sudo /bin/bash < /scripts/istio.sh",
		}, os.Stdout)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("deploying Istio service mesh failed")
		}
	}

	if prometheusEnabled {
		nodeName := nodeNameFromIndex(1)

		params := fmt.Sprintf("--prometheus-replicas %s ", strconv.Itoa(int(prometheusReplicas)))

		if prometheusAlertmanagerEnabled {
			params += fmt.Sprintf("--alertmanager true --alertmanager-replicas %s ", strconv.Itoa(int(prometheusAlertmanagerReplicas)))
		}

		if grafanaEnabled {
			params += fmt.Sprintf("--grafana true ")
		}

		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{
			"/bin/bash",
			"-c",
			fmt.Sprintf("ssh.sh sudo /bin/bash -s -- %s < /scripts/prometheus.sh", params),
		}, os.Stdout)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("deploying Prometheus operator failed")
		}
	}

	// If logging is enabled, deploy the default fluent logging
	if logDir != "" {
		nodeName := nodeNameFromIndex(1)
		success, err := docker.Exec(cli, nodeContainer(prefix, nodeName), []string{
			"/bin/bash",
			"-c",
			"ssh.sh sudo /bin/bash < /scripts/logging.sh",
		}, os.Stdout)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("provisioning logging failed")
		}
	}

	// If background flag was specified, we don't want to clean up if we reach that state
	if !background {
		wg.Wait()
		stop <- fmt.Errorf("Done. please clean up")
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

// prepareDeviceForAssignment moves the deivce from it's original driver to vfio-pci driver
func prepareDeviceForAssignment(cli *client.Client, nodeContainer, pciID, pciAddress string) error {
	devicePCIID := pciID
	if pciAddress != "" {
		devicePCIID, _ = getDevicePCIID(pciAddress)
	}
	success, err := docker.Exec(cli, nodeContainer, []string{
		"/bin/bash",
		"-c",
		fmt.Sprintf("ssh.sh sudo /bin/bash -s -- --vendor %s < /scripts/bind_device_to_vfio.sh", devicePCIID),
	}, os.Stdout)
	if err != nil {
		return err
	}
	if !success {
		return fmt.Errorf("binding device to vfio driver failed")
	}
	return nil
}

func prepareEtcdDataMount(node string, etcdDataDir string, mountSize string) error {
	var err error
	var success bool

	success, err = docker.Exec(cli, node, []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo mkdir -p %s", etcdDataDir)}, os.Stdout)
	if !success || err != nil {
		return fmt.Errorf("create etcd data directory '%s'on node %s failed: %v", etcdDataDir, node, err)
	}

	success, err = docker.Exec(cli, node, []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo test -d %s", etcdDataDir)}, os.Stdout)
	if !success || err != nil {
		return fmt.Errorf("verify etcd data directory '%s'on node %s exists failed: %v", etcdDataDir, node, err)
	}

	success, err = docker.Exec(cli, node, []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo mount -t tmpfs -o size=%s tmpfs %s", mountSize, etcdDataDir)}, os.Stdout)
	if !success || err != nil {
		return fmt.Errorf("create tmpfs mount '%s' for etcd data on node %s failed: %v", etcdDataDir, node, err)
	}

	success, err = docker.Exec(cli, node, []string{"/bin/bash", "-c", fmt.Sprintf("ssh.sh sudo df -h %s", etcdDataDir)}, os.Stdout)
	if !success || err != nil {
		return fmt.Errorf("verify that a mount for etcd data is exists on node %s failed: %v", node, err)
	}

	return nil
}
