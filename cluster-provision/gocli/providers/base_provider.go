package providers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alessio/shellescape"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	bindvfio "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/bind-vfio"
	dockerproxy "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/docker-proxy"
	etcdinmemory "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/etcd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/node01"
	nodeprovisioner "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/nodes"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/psa"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/realtime"
	sshutils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

func NewKubevirtProvider(k8sversion string, image string, cli *client.Client, options ...KubevirtProviderOption) *KubevirtProvider {
	bp := &KubevirtProvider{
		Version:     k8sversion,
		Nodes:       1,
		Numa:        1,
		Memory:      "3096M",
		CPU:         2,
		Background:  true,
		Image:       image,
		RandomPorts: true,
		Docker:      cli,
	}

	for _, option := range options {
		option(bp)
	}

	return bp
}

func (kp *KubevirtProvider) Start(ctx context.Context, cancel context.CancelFunc, portMap nat.PortMap) (retErr error) {
	stop := make(chan error, 10)
	containers, _, done := docker.NewCleanupHandler(kp.Docker, stop, os.Stdout, false)

	defer func() {
		stop <- retErr
		<-done
	}()

	go kp.handleInterrupt(cancel, stop)

	dnsmasq, err := kp.runDNSMasq(ctx, portMap)
	if err != nil {
		return err
	}
	kp.DNSMasq = dnsmasq
	containers <- dnsmasq

	dnsmasqJSON, err := kp.Docker.ContainerInspect(context.Background(), kp.DNSMasq)
	if err != nil {
		return err
	}

	if kp.SSHPort == 0 {
		port, err := utils.GetPublicPort(utils.PortSSH, dnsmasqJSON.NetworkSettings.Ports)
		if err != nil {
			return err
		}
		kp.SSHPort = port
	}

	if kp.APIServerPort == 0 {
		port, err := utils.GetPublicPort(utils.PortAPI, dnsmasqJSON.NetworkSettings.Ports)
		if err != nil {
			return err
		}
		kp.APIServerPort = port
	}
	_, err = utils.GetPublicPort(utils.PortAPI, dnsmasqJSON.NetworkSettings.Ports)

	registry, err := kp.runRegistry(ctx)
	if err != nil {
		return err
	}
	containers <- registry

	if kp.NFSData != "" {
		nfsGanesha, err := kp.runNFSGanesha(ctx)
		if err != nil {
			return nil
		}
		containers <- nfsGanesha
	}

	nodeIds, err := kp.runNodes(ctx)
	if err != nil {
		return err
	}
	for _, node := range nodeIds {
		containers <- node
	}

	return nil
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

	if err := kp.Docker.ContainerStart(ctx, dnsmasq.ID, types.ContainerStartOptions{}); err != nil {
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

	if err := kp.Docker.ContainerStart(ctx, registry.ID, types.ContainerStartOptions{}); err != nil {
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

	if err := kp.Docker.ContainerStart(ctx, nfsGanesha.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	return nfsGanesha.ID, nil
}

func (kp *KubevirtProvider) runNodes(ctx context.Context) ([]string, error) {
	wg := sync.WaitGroup{}
	wg.Add(int(kp.Nodes))
	containerIDs := []string{}

	// start one vm after each other
	macCounter := 0

	for x := 0; x < int(kp.Nodes); x++ {
		nodeName := kp.nodeNameFromIndex(x + 1)
		nodeNum := fmt.Sprintf("%02d", x+1)
		qemuCMD := kp.prepareQemuCmd(x)
		macCounter++

		vmContainerConfig := &container.Config{
			Image: kp.Image,
			Env: []string{
				fmt.Sprintf("NODE_NUM=%s", nodeNum),
			},
			Cmd: []string{"/bin/bash", "-c", qemuCMD},
		}

		deviceMappings, err := kp.prepareDeviceMappings()
		if err != nil {
			return nil, err
		}

		if kp.EnableCeph {
			vmContainerConfig.Volumes = map[string]struct{}{
				"/var/lib/rook": {},
			}
		}

		node, err := kp.Docker.ContainerCreate(ctx, vmContainerConfig, &container.HostConfig{
			Privileged:  true,
			NetworkMode: container.NetworkMode("container:" + kp.DNSMasq),
			Resources: container.Resources{
				Devices: deviceMappings,
			},
		}, nil, nil, kp.Version+"-"+nodeName)
		if err != nil {
			return nil, err
		}
		containerIDs = append(containerIDs, node.ID)

		if err := kp.Docker.ContainerStart(ctx, node.ID, types.ContainerStartOptions{}); err != nil {
			return nil, err
		}

		// Wait for vm start
		success, err := docker.Exec(kp.Docker, kp.nodeContainer(kp.Version, nodeName), []string{"/bin/bash", "-c", "while [ ! -f /ssh_ready ] ; do sleep 1; done"}, os.Stdout)
		if err != nil {
			return nil, err
		}

		if !success {
			return nil, fmt.Errorf("checking for ssh.sh script for node %s failed", nodeName)
		}

		err = kp.waitForVMToBeUp(kp.Version, nodeName)
		if err != nil {
			return nil, err
		}

		// turn to opt
		if kp.EnableFIPS {
			if _, err := sshutils.JumpSSH(kp.SSHPort, 1, "fips-mode-setup --enable && ( ssh.sh sudo reboot || true )", true, true); err != nil {
				return nil, err
			}
			err = kp.waitForVMToBeUp(kp.Version, nodeName)
			if err != nil {
				return nil, err
			}
		}

		// turn to opt
		if kp.DockerProxy != "" {
			//if dockerProxy has value, generate a shell script`/script/docker-proxy.sh` which can be applied to set proxy settings
			proxyOpt := dockerproxy.NewDockerProxyOpt(kp.SSHPort, kp.DockerProxy, x)
			if err := proxyOpt.Exec(); err != nil {
				return nil, err
			}
		}

		// turn to opt
		if kp.RunEtcdOnMemory {
			logrus.Infof("Creating in-memory mount for etcd data on node %s", nodeName)
			etcdOpt := etcdinmemory.NewEtcdInMemOpt(kp.SSHPort, x, kp.EtcdCapacity)
			if err = etcdOpt.Exec(); err != nil {
				return nil, err
			}
		}

		if kp.EnableRealtimeScheduler {
			realtimeOpt := realtime.NewRealtimeOpt(kp.SSHPort, x+1)
			if err := realtimeOpt.Exec(); err != nil {
				panic(err)
			}
		}

		//check if we have a special provision script
		success, err = docker.Exec(kp.Docker, kp.nodeContainer(kp.Version, nodeName), []string{"/bin/bash", "-c", fmt.Sprintf("test -f /scripts/%s.sh", nodeName)}, os.Stdout)
		if err != nil {
			return nil, fmt.Errorf("checking for matching provision script for node %s failed", nodeName)
		}
		// turn to opt
		for _, s := range []string{"8086:2668", "8086:2415"} {
			// move the VM sound cards to a vfio-pci driver to prepare for assignment
			bindVfioOpt := bindvfio.NewBindVfioOpt(kp.SSHPort, x, s)
			if err := bindVfioOpt.Exec(); err != nil {
				return nil, err
			}
			// turn to opt
			// err = prepareDeviceForAssignment(kp.Docker, kp.nodeContainer(kp.Version, nodeName), s, "")
			// if err != nil {
			// 	return nil, err
			// }
		}

		// turn to opt
		if kp.SingleStack {
			if _, err := sshutils.JumpSSH(kp.SSHPort, 1, "touch /home/vagrant/single_stack", true, true); err != nil {
				return nil, err
			}
		}

		// turn to opt
		if kp.EnableAudit {
			if _, err := sshutils.JumpSSH(kp.SSHPort, 1, "touch /home/vagrant/enable_audit", true, true); err != nil {
				return nil, err
			}
		}

		// turn to opt
		if kp.EnablePSA {
			psaOpt := psa.NewPsaOpt(kp.SSHPort)
			if err := psaOpt.Exec(); err != nil {
				return nil, err
			}
		}
		// todo: remove checking for scripts for node, just do different stuff at index 1
		if success {
			n := node01.NewNode01Provisioner(uint16(kp.SSHPort))
			err := n.Exec()
			if err != nil {
				panic(err)
			}
		} else {
			if kp.GPU != "" {
				// move the assigned PCI device to a vfio-pci driver to prepare for assignment
				// turn to opt
				// err = kp.prepareDeviceForAssignment(kp.Docker, kp.nodeContainer(kp.Version, nodeName), "", kp.GPU)
				// if err != nil {
				// 	return nil, err
				// }
			}
			n := nodeprovisioner.NewNodesProvisioner(kp.SSHPort, x+1)
			err = n.Exec()
			if err != nil {
				panic(err)
			}
		}

		if err != nil {
			return nil, err
		}

		go func(id string) {
			kp.Docker.ContainerWait(ctx, id, container.WaitConditionNotRunning)
			wg.Done()
		}(node.ID)
	}

	return containerIDs, nil
}

func (kp *KubevirtProvider) prepareDeviceMappings() ([]container.DeviceMapping, error) {
	iommu_group, err := kp.getPCIDeviceIOMMUGroup(kp.GPU)
	if err != nil {
		return nil, err
	}
	vfioDevice := fmt.Sprintf("/dev/vfio/%s", iommu_group)
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

// func (kp *KubevirtProvider) Stop() {}

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
