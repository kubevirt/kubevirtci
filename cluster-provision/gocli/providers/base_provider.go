package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts"
	aaq "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/aaq"
	bindvfio "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/bind-vfio"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/cdi"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/cnao"
	dockerproxy "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/docker-proxy"
	etcd "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/etcd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/istio"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/k8sprovision"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/ksm"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/labelnodes"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/multus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/nfscsi"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/node01"
	nodesprovision "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/nodes"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/prometheus"
	provisionopt "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/provision"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/psa"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/realtime"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/rookceph"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/rootkey"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/swap"
	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/k8s"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
)

var versionMap = map[string]string{
	"1.30": "1.30.2",
	"1.29": "1.29.6",
	"1.28": "1.28.11",
}

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

	wg := sync.WaitGroup{}
	wg.Add(int(kp.Nodes))
	macCounter := 0

	for x := 0; x < int(kp.Nodes); x++ {
		nodeName := kp.nodeNameFromIndex(x + 1)
		sshClient, err := libssh.NewSSHClient(kp.SSHPort, x+1, false)
		if err != nil {
			return err
		}

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
		var deviceMappings []container.DeviceMapping

		if kp.GPU != "" && x == int(kp.Nodes)-1 {
			dm, err := kp.prepareDeviceMappings()
			if err != nil {
				return err
			}
			deviceMappings = dm
			qemuCMD = fmt.Sprintf("%s -device vfio-pci,host=%s", qemuCMD, kp.GPU)
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
			return err
		}
		containers <- node.ID

		if err := kp.Docker.ContainerStart(ctx, node.ID, types.ContainerStartOptions{}); err != nil {
			return err
		}

		success, err := docker.Exec(kp.Docker, kp.nodeContainer(kp.Version, nodeName), []string{"/bin/bash", "-c", "while [ ! -f /ssh_ready ] ; do sleep 1; done"}, os.Stdout)
		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("checking for ssh.sh script for node %s failed", nodeName)
		}

		err = kp.waitForVMToBeUp(kp.Version, nodeName)
		if err != nil {
			return err
		}

		rootkey := rootkey.NewRootKey(sshClient)
		if err = rootkey.Exec(); err != nil {
			return err
		}
		sshClient, err = libssh.NewSSHClient(kp.SSHPort, x+1, true)

		if err = kp.provisionNode(sshClient, x+1); err != nil {
			return err
		}

		go func(id string) {
			kp.Docker.ContainerWait(ctx, id, container.WaitConditionNotRunning)
			wg.Done()
		}(node.ID)
	}

	sshClient, err := libssh.NewSSHClient(kp.SSHPort, 1, true)
	if err != nil {
		return err
	}

	err = sshClient.CopyRemoteFile("/etc/kubernetes/admin.conf", ".kubeconfig")
	if err != nil {
		return err
	}

	config, err := k8s.InitConfig(".kubeconfig", kp.APIServerPort)
	if err != nil {
		return err
	}

	k8sClient, err := k8s.NewDynamicClient(config)
	if err != nil {
		return err
	}
	kp.Client = k8sClient

	if err = kp.provisionK8sOpts(sshClient); err != nil {
		return err
	}

	err = kp.persistProvider()
	if err != nil {
		return err
	}

	return nil
}

func (kp *KubevirtProvider) Provision(ctx context.Context, cancel context.CancelFunc, portMap nat.PortMap) (retErr error) {
	prefix := fmt.Sprintf("k8s-%s-provision", kp.Version)
	target := fmt.Sprintf("quay.io/kubevirtci/k8s-%s", kp.Version)
	if kp.Phases == "linux" {
		target = kp.Image + "-base"
	}
	version := kp.Version
	kp.Version = prefix

	stop := make(chan error, 10)
	containers, volumes, done := docker.NewCleanupHandler(kp.Docker, stop, os.Stdout, true)

	defer func() {
		stop <- retErr
		<-done
	}()

	go kp.handleInterrupt(cancel, stop)

	err := docker.ImagePull(kp.Docker, ctx, kp.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}

	dnsmasq, err := kp.runDNSMasq(ctx, portMap)
	if err != nil {
		return err
	}

	kp.DNSMasq = dnsmasq
	containers <- dnsmasq

	dm, err := kp.Docker.ContainerInspect(context.Background(), dnsmasq)
	if err != nil {
		return err
	}

	sshPort, err := utils.GetPublicPort(utils.PortSSH, dm.NetworkSettings.Ports)
	if err != nil {
		return err
	}

	nodeName := kp.nodeNameFromIndex(1)
	nodeNum := fmt.Sprintf("%02d", 1)

	vol, err := kp.Docker.VolumeCreate(ctx, volume.CreateOptions{
		Name: fmt.Sprintf("%s-%s", prefix, nodeName),
	})
	if err != nil {
		return err
	}
	volumes <- vol.Name
	registryVol, err := kp.Docker.VolumeCreate(ctx, volume.CreateOptions{
		Name: fmt.Sprintf("%s-%s", prefix, "registry"),
	})
	if err != nil {
		return err
	}

	node, err := kp.Docker.ContainerCreate(ctx, &container.Config{
		Image: kp.Image,
		Env: []string{
			fmt.Sprintf("NODE_NUM=%s", nodeNum),
		},
		Volumes: map[string]struct{}{
			"/var/run/disk":     {},
			"/var/lib/registry": {},
		},
		Cmd: []string{"/bin/bash", "-c", fmt.Sprintf("/vm.sh --memory %s --cpu %s %s", kp.Memory, strconv.Itoa(int(kp.CPU)), kp.QemuArgs)},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "volume",
				Source: vol.Name,
				Target: "/var/run/disk",
			},
			{
				Type:   "volume",
				Source: registryVol.Name,
				Target: "/var/lib/registry",
			},
		},
		Privileged:  true,
		NetworkMode: container.NetworkMode("container:" + kp.DNSMasq),
	}, nil, nil, kp.nodeContainer(kp.Version, nodeName))
	if err != nil {
		return err
	}
	containers <- node.ID
	if err := kp.Docker.ContainerStart(ctx, node.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// Wait for ssh.sh script to exist
	_, err = docker.Exec(kp.Docker, kp.nodeContainer(kp.Version, nodeName), []string{"bin/bash", "-c", "while [ ! -f /ssh_ready ] ; do sleep 1; done", "checking for ssh.sh script"}, os.Stdout)
	if err != nil {
		return err
	}

	sshClient, err := libssh.NewSSHClient(sshPort, 1, false)
	if err != nil {
		return err
	}

	rootkey := rootkey.NewRootKey(sshClient)
	if err = rootkey.Exec(); err != nil {
		fmt.Println(err)
	}

	sshClient, err = libssh.NewSSHClient(sshPort, 1, true)
	if err != nil {
		return err
	}

	if strings.Contains(kp.Phases, "linux") {
		provisionOpt := provisionopt.NewLinuxProvisioner(sshClient)
		if err = provisionOpt.Exec(); err != nil {
			return err
		}
	}

	if strings.Contains(kp.Phases, "k8s") {
		// copy provider scripts
		if _, err = sshClient.Command("mkdir -p /tmp/ceph /tmp/cnao /tmp/nfs-csi /tmp/nodeports /tmp/prometheus /tmp/whereabouts /tmp/kwok", true); err != nil {
			return err
		}
		// Copy manifests to the VM
		success, err := docker.Exec(kp.Docker, kp.nodeContainer(kp.Version, nodeName), []string{"/bin/bash", "-c", "scp -r -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i vagrant.key -P 22 /scripts/manifests/* root@192.168.66.101:/tmp"}, os.Stdout)
		if err != nil {
			return err
		}

		if !success {
			return fmt.Errorf("error copying manifests to node")
		}

		versionWithMinor, ok := versionMap[version]
		if !ok {
			return fmt.Errorf("Invalid version")
		}

		provisionK8sOpt := k8sprovision.NewK8sProvisioner(sshClient, versionWithMinor, kp.Slim)
		if err = provisionK8sOpt.Exec(); err != nil {
			return err
		}
	}

	if _, err = sshClient.Command("sudo shutdown now -h", true); err != nil {
		return err
	}

	_, err = docker.Exec(kp.Docker, kp.nodeContainer(kp.Version, nodeName), []string{"rm", "/ssh_ready"}, io.Discard)
	if err != nil {
		return err
	}

	logrus.Info("waiting for the node to stop")
	okChan, errChan := kp.Docker.ContainerWait(ctx, kp.nodeContainer(kp.Version, nodeName), container.WaitConditionNotRunning)
	select {
	case <-okChan:
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("waiting for the node to stop failed: %v", err)
		}
	}

	if len(kp.AdditionalKernelArgs) > 0 {
		dir, err := os.MkdirTemp("", "gocli")
		if err != nil {
			return fmt.Errorf("failed creating a temporary directory: %v", err)
		}
		defer os.RemoveAll(dir)
		if err := os.WriteFile(filepath.Join(dir, "additional.kernel.args"), []byte(shellescape.QuoteCommand(kp.AdditionalKernelArgs)), 0666); err != nil {
			return fmt.Errorf("failed creating additional.kernel.args file: %v", err)
		}
		if err := kp.copyDirectory(ctx, kp.Docker, node.ID, dir, "/"); err != nil {
			return fmt.Errorf("failed copying additional kernel arguments into the container: %v", err)
		}
	}

	logrus.Infof("Commiting the node as %s", target)
	_, err = kp.Docker.ContainerCommit(ctx, node.ID, types.ContainerCommitOptions{
		Reference: target,
		Comment:   "PROVISION SUCCEEDED",
		Author:    "gocli",
		Changes:   nil,
		Pause:     false,
		Config:    nil,
	})
	if err != nil {
		return fmt.Errorf("commiting the node failed: %v", err)
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

func (kp *KubevirtProvider) provisionNode(sshClient libssh.Client, nodeIdx int) error {
	opts := []opts.Opt{}
	nodeName := kp.nodeNameFromIndex(nodeIdx)

	if kp.EnableFIPS {
		for _, cmd := range []string{"sudo fips-mode-setup --enable", "sudo reboot"} {
			if _, err := sshClient.Command(cmd, true); err != nil {
				return fmt.Errorf("Starting fips mode failed: %s", err)
			}
		}
		err := kp.waitForVMToBeUp(kp.Version, nodeName)
		if err != nil {
			return err
		}
	}

	if kp.DockerProxy != "" {
		//if dockerProxy has value, generate a shell script`/script/docker-proxy.sh` which can be applied to set proxy settings
		dp := dockerproxy.NewDockerProxyOpt(sshClient, kp.DockerProxy)
		opts = append(opts, dp)
	}

	if kp.RunEtcdOnMemory {
		logrus.Infof("Creating in-memory mount for etcd data on node %s", nodeName)
		etcdinmem := etcd.NewEtcdInMemOpt(sshClient, kp.EtcdCapacity)
		opts = append(opts, etcdinmem)
	}

	if kp.EnableRealtimeScheduler {
		realtimeOpt := realtime.NewRealtimeOpt(sshClient)
		opts = append(opts, realtimeOpt)
	}

	for _, s := range []string{"8086:2668", "8086:2415"} {
		// move the VM sound cards to a vfio-pci driver to prepare for assignment
		bvfio := bindvfio.NewBindVfioOpt(sshClient, s)
		opts = append(opts, bvfio)
	}

	if kp.SingleStack {
		if _, err := sshClient.Command("touch /home/vagrant/single_stack", false); err != nil {
			return fmt.Errorf("provisioning node %d failed (setting singleStack phase): %s", nodeIdx, err)
		}
	}

	if kp.EnableAudit {
		if _, err := sshClient.Command("touch /home/vagrant/enable_audit", false); err != nil {
			return fmt.Errorf("provisioning node %d failed (setting enableAudit phase): %s", nodeIdx, err)
		}
	}

	if kp.EnablePSA {
		psaOpt := psa.NewPsaOpt(sshClient)
		opts = append(opts, psaOpt)
	}

	if nodeIdx == 1 {
		n := node01.NewNode01Provisioner(sshClient)
		opts = append(opts, n)

	} else {
		if kp.GPU != "" {
			// move the assigned PCI device to a vfio-pci driver to prepare for assignment
			gpuDeviceID, err := kp.getDevicePCIID(kp.GPU)
			if err != nil {
				return err
			}
			bindVfioOpt := bindvfio.NewBindVfioOpt(sshClient, gpuDeviceID)
			opts = append(opts, bindVfioOpt)
		}
		n := nodesprovision.NewNodesProvisioner(sshClient)
		opts = append(opts, n)
	}

	if kp.KSM {
		ksmOpt := ksm.NewKsmOpt(sshClient, int(kp.KSMInterval), int(kp.KSMPages))
		opts = append(opts, ksmOpt)
	}

	if kp.Swap {
		swapOpt := swap.NewSwapOpt(sshClient, int(kp.Swapiness), kp.UnlimitedSwap, int(kp.Swapsize))
		opts = append(opts, swapOpt)
	}

	for _, o := range opts {
		if err := o.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (kp *KubevirtProvider) provisionK8sOpts(sshClient libssh.Client) error {
	opts := []opts.Opt{}
	labelSelector := "node-role.kubernetes.io/control-plane"
	if kp.Nodes > 1 {
		labelSelector = "!node-role.kubernetes.io/control-plane"
	}
	opts = append(opts, labelnodes.NewNodeLabler(sshClient, labelSelector))

	if kp.CDI {
		opts = append(opts, cdi.NewCdiOpt(kp.Client, kp.CDIVersion))
	}

	if kp.AAQ {
		if kp.Version == "k8s-1.30" {
			opts = append(opts, aaq.NewAaqOpt(kp.Client, sshClient, kp.AAQVersion))
		} else {
			logrus.Info("AAQ was requested but kubernetes version is less than 1.30, skipping")
		}
	}

	if kp.EnablePrometheus {
		opts = append(opts, prometheus.NewPrometheusOpt(kp.Client, kp.EnableGrafana, kp.EnablePrometheusAlertManager))
	}

	if kp.EnableCeph {
		opts = append(opts, rookceph.NewCephOpt(kp.Client))
	}

	if kp.EnableNFSCSI {
		opts = append(opts, nfscsi.NewNfsCsiOpt(kp.Client))
	}

	if kp.EnableMultus {
		opts = append(opts, multus.NewMultusOpt(kp.Client, sshClient))
	}

	if kp.EnableCNAO {
		opts = append(opts, cnao.NewCnaoOpt(kp.Client, sshClient))
	}

	if kp.EnableIstio {
		opts = append(opts, istio.NewIstioOpt(sshClient, kp.Client, kp.EnableCNAO))
	}

	for _, opt := range opts {
		if err := opt.Exec(); err != nil {
			return err
		}
	}

	return nil
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

func (kp *KubevirtProvider) getPCIDeviceIOMMUGroup(address string) (string, error) {
	iommuLink := filepath.Join("/sys/bus/pci/devices", address, "iommu_group")
	iommuPath, err := os.Readlink(iommuLink)
	if err != nil {
		return "", fmt.Errorf("failed to read iommu_group link %s for device %s - %v", iommuLink, address, err)
	}
	_, iommuGroup := filepath.Split(iommuPath)
	return iommuGroup, nil
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
