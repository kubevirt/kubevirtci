package providers

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/aaq"
	bindvfio "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/bind-vfio"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/cdi"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/cnao"
	dockerproxy "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/docker-proxy"
	etcd "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/etcd"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/istio"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/ksm"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/labelnodes"
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
)

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
		kp.SSHPort = uint(port)
	}

	if kp.APIServerPort == 0 {
		port, err := utils.GetPublicPort(utils.PortAPI, dnsmasqJSON.NetworkSettings.Ports)
		if err != nil {
			return err
		}
		kp.APIServerPort = uint(port)
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
		sshClient, err := libssh.NewSSHClient(uint16(kp.SSHPort), x+1, false)
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

		if err := kp.Docker.ContainerStart(ctx, node.ID, container.StartOptions{}); err != nil {
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
		sshClient, err = libssh.NewSSHClient(uint16(kp.SSHPort), x+1, true)

		if err = kp.provisionNode(sshClient, x+1); err != nil {
			return err
		}

		go func(id string) {
			kp.Docker.ContainerWait(ctx, id, container.WaitConditionNotRunning)
			wg.Done()
		}(node.ID)
	}

	sshClient, err := libssh.NewSSHClient(uint16(kp.SSHPort), 1, true)
	if err != nil {
		return err
	}

	kubeConf, err := os.Create(".kubeconfig")
	if err != nil {
		return err
	}

	err = sshClient.CopyRemoteFile("/etc/kubernetes/admin.conf", kubeConf)
	if err != nil {
		return err
	}

	config, err := k8s.NewConfig(".kubeconfig", uint16(kp.APIServerPort))
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

func (kp *KubevirtProvider) provisionNode(sshClient libssh.Client, nodeIdx int) error {
	opts := []opts.Opt{}
	nodeName := kp.nodeNameFromIndex(nodeIdx)

	if kp.EnableFIPS {
		for _, cmd := range []string{"sudo fips-mode-setup --enable", "sudo reboot"} {
			if err := sshClient.Command(cmd); err != nil {
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

	if kp.EnableAudit {
		if err := sshClient.Command("touch /home/vagrant/enable_audit"); err != nil {
			return fmt.Errorf("provisioning node %d failed (setting enableAudit phase): %s", nodeIdx, err)
		}
	}

	if kp.EnablePSA {
		psaOpt := psa.NewPsaOpt(sshClient)
		opts = append(opts, psaOpt)
	}

	if nodeIdx == 1 {
		n := node01.NewNode01Provisioner(sshClient, kp.SingleStack, kp.NoEtcdFsync)
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
		n := nodesprovision.NewNodesProvisioner(sshClient, kp.SingleStack)
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
		opts = append(opts, cdi.NewCdiOpt(kp.Client, sshClient, kp.CDIVersion))
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
		opts = append(opts, rookceph.NewCephOpt(kp.Client, sshClient))
	}

	if kp.EnableNFSCSI {
		opts = append(opts, nfscsi.NewNfsCsiOpt(kp.Client))
	}

	if kp.EnableMultus {
		opts = append(opts, multus.NewMultusOpt(kp.Client, sshClient))
	}

	if kp.EnableCNAO {
		opts = append(opts, cnao.NewCnaoOpt(kp.Client, sshClient, kp.EnableMultus, kp.SkipCnaoCR))
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
