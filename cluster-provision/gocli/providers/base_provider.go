package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alessio/shellescape"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

func NewKubevirtProvider(k8sversion string, image string, cli *client.Client,
	options []KubevirtProviderOption) *KubevirtProvider {
	kp := &KubevirtProvider{
		Image:   image,
		Version: k8sversion,
		Docker:  cli,
		Nodes:   1, // start with nodes equal one and will be later modified by options that set a different value
	}

	for _, option := range options {
		option(kp)
	}

	return kp
}

func NewFromRunning(dnsmasqPrefix string) (*KubevirtProvider, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	containers, err := docker.GetPrefixedContainers(cli, dnsmasqPrefix+"-dnsmasq")
	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, fmt.Errorf("No running provider has the prefix %s", dnsmasqPrefix)
	}

	var buf bytes.Buffer
	_, err = docker.Exec(cli, containers[0].ID, []string{"cat", "provider.json"}, &buf)
	if err != nil {
		return nil, err
	}

	kp := &KubevirtProvider{}

	err = json.Unmarshal(buf.Bytes(), kp)
	if err != nil {
		return nil, err
	}

	kp.Docker = cli
	return kp, nil
}

func (kp *KubevirtProvider) runDNSMasq(ctx context.Context, portMap nat.PortMap) (string, error) {
	dnsmasqMounts := []mount.Mount{}
	_, err := os.Stat("/lib/modules")
	if err == nil {
		dnsmasqMounts = []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/lib/modules",
				Target: "/lib/modules",
			},
		}
	}

	dnsmasq, err := kp.Docker.ContainerCreate(ctx, &container.Config{
		Image: kp.Image,
		Env: []string{
			fmt.Sprintf("NUM_NODES=%d", kp.Nodes),
			fmt.Sprintf("NUM_SECONDARY_NICS=%d", kp.SecondaryNics),
		},
		Cmd: []string{"/bin/bash", "-c", "/dnsmasq.sh"},
		ExposedPorts: nat.PortSet{
			utils.TCPPortOrDie(utils.PortSSH):         {},
			utils.TCPPortOrDie(utils.PortRegistry):    {},
			utils.TCPPortOrDie(utils.PortOCP):         {},
			utils.TCPPortOrDie(utils.PortAPI):         {},
			utils.TCPPortOrDie(utils.PortVNC):         {},
			utils.TCPPortOrDie(utils.PortHTTP):        {},
			utils.TCPPortOrDie(utils.PortHTTPS):       {},
			utils.TCPPortOrDie(utils.PortPrometheus):  {},
			utils.TCPPortOrDie(utils.PortGrafana):     {},
			utils.TCPPortOrDie(utils.PortUploadProxy): {},
			utils.UDPPortOrDie(utils.PortDNS):         {},
		},
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: kp.RandomPorts,
		PortBindings:    portMap,
		ExtraHosts: []string{
			"nfs:192.168.66.2",
			"registry:192.168.66.2",
			"ceph:192.168.66.2",
		},
		Mounts: dnsmasqMounts,
	}, nil, nil, kp.Version+"-dnsmasq")

	if err := kp.Docker.ContainerStart(ctx, dnsmasq.ID, container.StartOptions{}); err != nil {
		return "", err
	}
	return dnsmasq.ID, nil
}

func (kp *KubevirtProvider) runRegistry(ctx context.Context) (string, error) {
	err := docker.ImagePull(kp.Docker, ctx, utils.DockerRegistryImage, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	registry, err := kp.Docker.ContainerCreate(ctx, &container.Config{
		Image: utils.DockerRegistryImage,
	}, &container.HostConfig{
		Privileged:  true,
		NetworkMode: container.NetworkMode("container:" + kp.DNSMasq),
	}, nil, nil, kp.Version+"-registry")
	if err != nil {
		return "", err
	}

	if err := kp.Docker.ContainerStart(ctx, registry.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return registry.ID, nil
}

func (kp *KubevirtProvider) runNFSGanesha(ctx context.Context) (string, error) {
	nfsData, err := filepath.Abs(kp.NFSData)
	if err != nil {
		return "", err
	}

	err = docker.ImagePull(kp.Docker, ctx, utils.NFSGaneshaImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	nfsGanesha, err := kp.Docker.ContainerCreate(ctx, &container.Config{
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
		NetworkMode: container.NetworkMode("container:" + kp.DNSMasq),
	}, nil, nil, kp.Version+"-nfs-ganesha")
	if err != nil {
		return "", err
	}

	if err := kp.Docker.ContainerStart(ctx, nfsGanesha.ID, container.StartOptions{}); err != nil {
		return "", err
	}
	return nfsGanesha.ID, nil
}

func (kp *KubevirtProvider) prepareDeviceMappings() ([]container.DeviceMapping, error) {
	iommuGroup, err := kp.getPCIDeviceIOMMUGroup(kp.GPU)
	if err != nil {
		return nil, err
	}
	vfioDevice := fmt.Sprintf("/dev/vfio/%s", iommuGroup)
	return []container.DeviceMapping{
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
	}, nil
}

func (kp *KubevirtProvider) prepareQemuCmd(x int) string {
	nodeQemuArgs := kp.QemuArgs
	kernelArgs := kp.KernelArgs
	macSuffix := fmt.Sprintf("%02x", x)

	for i := 0; i < int(kp.SecondaryNics); i++ {
		netSuffix := fmt.Sprintf("%d-%d", x, i)
		nodeQemuArgs = fmt.Sprintf("%s -device virtio-net-pci,netdev=secondarynet%s,mac=52:55:00:d1:56:%s -netdev tap,id=secondarynet%s,ifname=stap%s,script=no,downscript=no", nodeQemuArgs, netSuffix, macSuffix, netSuffix, netSuffix)
	}

	if kp.GPU != "" && x == int(kp.Nodes)-1 {
		nodeQemuArgs = fmt.Sprintf("%s -device vfio-pci,host=%s", nodeQemuArgs, kp.GPU)
	}

	var vmArgsNvmeDisks []string
	if len(kp.NvmeDisks) > 0 {
		for i, size := range kp.NvmeDisks {
			resource.MustParse(size)
			disk := fmt.Sprintf("%s-%d.img", "/nvme", i)
			nodeQemuArgs = fmt.Sprintf("%s -drive file=%s,format=raw,id=NVME%d,if=none -device nvme,drive=NVME%d,serial=nvme-%d", nodeQemuArgs, disk, i, i, i)
			vmArgsNvmeDisks = append(vmArgsNvmeDisks, fmt.Sprintf("--nvme-device-size %s", size))
		}
	}
	var vmArgsSCSIDisks []string
	if len(kp.ScsiDisks) > 0 {
		nodeQemuArgs = fmt.Sprintf("%s -device virtio-scsi-pci,id=scsi0", nodeQemuArgs)
		for i, size := range kp.ScsiDisks {
			resource.MustParse(size)
			disk := fmt.Sprintf("%s-%d.img", "/scsi", i)
			nodeQemuArgs = fmt.Sprintf("%s -drive file=%s,if=none,id=drive%d -device scsi-hd,drive=drive%d,bus=scsi0.0,channel=0,scsi-id=0,lun=%d", nodeQemuArgs, disk, i, i, i)
			vmArgsSCSIDisks = append(vmArgsSCSIDisks, fmt.Sprintf("--scsi-device-size %s", size))
		}
	}

	var vmArgsUSBDisks []string
	const bus = " -device qemu-xhci,id=bus%d"
	const drive = " -drive if=none,id=stick%d,format=raw,file=/usb-%d.img"
	const dev = " -device usb-storage,bus=bus%d.0,drive=stick%d"
	const usbSizefmt = " --usb-device-size %s"
	if len(kp.USBDisks) > 0 {
		for i, size := range kp.USBDisks {
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

	if kp.Hugepages2M > 0 {
		kernelArgs += fmt.Sprintf(" hugepagesz=2M hugepages=%d", kp.Hugepages2M)
	}

	if kp.Hugepages1G > 0 {
		kernelArgs += fmt.Sprintf(" hugepagesz=1G hugepages=%d", kp.Hugepages1G)
	}

	if kp.EnableFIPS {
		kernelArgs += " fips=1"
	}

	blockDev := ""
	if kp.EnableCeph {
		blockDev = "--block-device /var/run/disk/blockdev.qcow2 --block-device-size 32212254720"
	}

	kernelArgs = strings.TrimSpace(kernelArgs)
	if kernelArgs != "" {
		additionalArgs = append(additionalArgs, "--additional-kernel-args", shellescape.Quote(kernelArgs))
	}

	return fmt.Sprintf("/vm.sh -n /var/run/disk/disk.qcow2 --memory %s --cpu %s --numa %s %s %s %s %s %s",
		kp.Memory,
		strconv.Itoa(int(kp.CPU)),
		strconv.Itoa(int(kp.Numa)),
		blockDev,
		strings.Join(vmArgsSCSIDisks, " "),
		strings.Join(vmArgsNvmeDisks, " "),
		strings.Join(vmArgsUSBDisks, " "),
		strings.Join(additionalArgs, " "),
	)
}

func (kp *KubevirtProvider) persistProvider() error {
	providerJson, err := json.Marshal(kp)
	if err != nil {
		return err
	}
	escapedJson := strconv.Quote(string(providerJson))

	_, err = docker.Exec(kp.Docker, kp.DNSMasq, []string{"/bin/bash", "-c", fmt.Sprintf("echo %s | tee /provider.json > /dev/null", string(escapedJson))}, os.Stdout)
	if err != nil {
		return err
	}
	return nil
}

func (kp *KubevirtProvider) getDevicePCIID(pciAddress string) (string, error) {
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

func (kp *KubevirtProvider) copyDirectory(ctx context.Context, cli *client.Client, containerID string, sourceDirectory string, targetDirectory string) error {
	srcInfo, err := archive.CopyInfoSourcePath(sourceDirectory, false)
	if err != nil {
		return err
	}

	srcArchive, err := archive.TarResource(srcInfo)
	if err != nil {
		return err
	}
	defer srcArchive.Close()

	dstInfo := archive.CopyInfo{Path: targetDirectory}

	dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
	if err != nil {
		return err
	}
	defer preparedArchive.Close()

	err = cli.CopyToContainer(ctx, containerID, dstDir, preparedArchive, types.CopyToContainerOptions{AllowOverwriteDirWithFile: false})
	if err != nil {
		return err
	}
	return nil
}

func (kp *KubevirtProvider) getPCIDeviceIOMMUGroup(address string) (string, error) {
	iommuLink := filepath.Join("/sys/bus/pci/devices", address, "iommu_group")
	iommuPath, err := os.Readlink(iommuLink)
	if err != nil {
		return "", fmt.Errorf("failed to read iommu_group link %s for device %s - %v", iommuLink, address, err)
	}
	_, iommuGroup := filepath.Split(iommuPath)
	return iommuGroup, nil
}

func (kp *KubevirtProvider) handleInterrupt(cancel context.CancelFunc, stop chan error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	cancel()
	stop <- fmt.Errorf("Interrupt received, clean up")
}

func (kp *KubevirtProvider) nodeNameFromIndex(x int) string {
	return fmt.Sprintf("node%02d", x)
}

func (kp *KubevirtProvider) nodeContainer(version string, node string) string {
	return version + "-" + node
}

func (kp *KubevirtProvider) waitForVMToBeUp(prefix string, nodeName string) error {
	var err error
	for x := 0; x < 10; x++ {
		_, err = docker.Exec(kp.Docker, kp.nodeContainer(prefix, nodeName), []string{"/bin/bash", "-c", "ssh.sh echo VM is up"}, os.Stdout)
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
