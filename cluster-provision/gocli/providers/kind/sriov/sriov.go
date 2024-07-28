package sriov

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/client"
	dockercri "kubevirt.io/kubevirtci/cluster-provision/gocli/cri/docker"
	podmancri "kubevirt.io/kubevirtci/cluster-provision/gocli/cri/podman"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
	multussriov "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/multus-sriov"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/opts/remountsysfs"
	sriovcomponents "kubevirt.io/kubevirtci/cluster-provision/gocli/opts/sriov-components"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/pkg/libssh"
	kind "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/kindbase"
)

type KindSriov struct {
	pfs            []string
	pfCountPerNode int
	vfsCount       int

	*kind.KindBaseProvider
}

func NewKindSriovProvider(kindConfig *kind.KindConfig) (*KindSriov, error) {
	kindBase, err := kind.NewKindBaseProvider(kindConfig)
	if err != nil {
		return nil, err
	}
	return &KindSriov{
		KindBaseProvider: kindBase,
	}, nil
}

func (ks *KindSriov) Start(ctx context.Context, cancel context.CancelFunc) error {
	devs, err := ks.discoverHostPFs()
	if err != nil {
		return err
	}
	ks.pfs = devs

	if ks.Nodes*ks.pfCountPerNode > len(devs) {
		return fmt.Errorf("Not enough virtual functions available, there are %d functions on the host", len(devs))
	}

	if err = ks.KindBaseProvider.Start(ctx, cancel); err != nil {
		return err
	}

	nodes, err := ks.Provider.ListNodes(ks.Version)
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	var sshClient, controlPlaneClient libssh.Client

	switch ks.CRI.(type) {
	case *dockercri.DockerClient:
		controlPlaneClient = docker.NewDockerAdapter(cli, ks.Version+"-control-plane")
	case *podmancri.Podman:
		controlPlaneClient = podmancri.NewPodmanSSHClient(ks.Version + "-control-plane")
	}

	pfOffset := 0
	for _, node := range nodes {
		nodeName := node.String()
		switch ks.CRI.(type) {
		case *dockercri.DockerClient:
			sshClient = docker.NewDockerAdapter(cli, node.String())
		case *podmancri.Podman:
			sshClient = podmancri.NewPodmanSSHClient(node.String())
		}

		resp, err := ks.CRI.Inspect(nodeName, "{{.State.Pid}}")
		if err != nil {
			return err
		}

		pid, err := strconv.Atoi(strings.TrimSuffix(string(resp), "\n"))
		if err != nil {
			return err
		}

		if err = ks.linkNetNS(pid, nodeName); err != nil {
			return err
		}

		pfsForNode := ks.pfs[pfOffset : pfOffset+ks.pfCountPerNode]
		if err = ks.assignPfsToNode(pfsForNode, nodeName); err != nil {
			return err
		}

		pfOffset += ks.pfCountPerNode

		rsf := remountsysfs.NewRemountSysFSOpt(sshClient)
		if err := rsf.Exec(); err != nil {
			return err
		}

		pfs, err := ks.fetchNodePfs(sshClient)
		if err != nil {
			return nil
		}

		for _, pf := range pfs {
			vfsSysFsDevices, err := ks.createVFsforPF(sshClient, pf)
			for _, vfDevice := range vfsSysFsDevices {
				err = ks.bindToVfio(sshClient, vfDevice)
				if err != nil {
					return err
				}
			}
		}

		if _, err = controlPlaneClient.Command("kubectl label node "+nodeName+" sriov_capable=true", true); err != nil {
			return err
		}
	}

	msrv := multussriov.NewMultusSriovOpt(ks.Client)
	if err = msrv.Exec(); err != nil {
		return err
	}

	components := sriovcomponents.NewSriovComponentsOpt(ks.Client)
	if err = components.Exec(); err != nil {
		return err
	}
	return nil
}

func (ks *KindSriov) discoverHostPFs() ([]string, error) {
	files, err := filepath.Glob("/sys/class/net/*/device/sriov_numvfs")
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, errors.New("FATAL: Could not find available sriov PFs on host")
	}

	pfNames := make([]string, 0)
	for _, file := range files {
		pfName := filepath.Base(filepath.Dir(filepath.Dir(file)))
		pfNames = append(pfNames, pfName)
	}

	return pfNames, nil
}

func (ks *KindSriov) assignPfsToNode(pfs []string, nodeName string) error {
	for _, pf := range pfs {
		cmds := []string{
			"link set " + pf + " netns " + nodeName,
			"netns exec " + nodeName + " ip link set up dev " + pf,
			"netns exec " + nodeName + " ip link show",
		}
		for _, cmd := range cmds {
			cmd := exec.Command("ip", cmd)
			if _, err := cmd.Output(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (ks *KindSriov) linkNetNS(pid int, nodeName string) error {
	cmd := exec.Command("ln", "-sf", "/proc/"+fmt.Sprintf("%d", pid)+"/ns/net", "/var/run/netns/"+nodeName)
	if _, err := cmd.CombinedOutput(); err != nil {
		return err
	}
	return nil
}

func (ks *KindSriov) fetchNodePfs(sshClient libssh.Client) ([]string, error) {
	mod, err := sshClient.Command(`grep vfio_pci /proc/modules`, false)
	if err != nil {
		return nil, err
	}
	if mod == "" {
		return nil, fmt.Errorf("System doesn't have the vfio_pci module, provisioning failed")
	}

	if _, err = sshClient.Command("modprobe -i vfio_pci", true); err != nil {
		return nil, err
	}

	pfsString, err := sshClient.Command(`find /sys/class/net/*/device/`, false)
	if err != nil {
		return nil, err
	}
	pfs := strings.Split(pfsString, "\n")
	if len(pfs) == 0 {
		return nil, fmt.Errorf("No physical functions found on node, exiting")
	}

	return pfs, nil
}

func (ks *KindSriov) createVFsforPF(sshClient libssh.Client, vfSysClassNetPath string) ([]string, error) {
	pfSysFsDevice, err := sshClient.Command("readlink -e "+vfSysClassNetPath, false)
	if err != nil {
		return nil, err
	}
	totalVfs, err := sshClient.Command("cat "+pfSysFsDevice+"/sriov_totalvfs", false)
	if err != nil {
		return nil, err
	}

	totalVfsCount, err := strconv.Atoi(totalVfs)
	if err != nil {
		return nil, err
	}

	if totalVfsCount < ks.vfsCount {
		return nil, fmt.Errorf("FATAL: PF %s, VF's count should be up to sriov_totalvfs: %d", vfSysClassNetPath, totalVfsCount)
	}

	cmds := []string{
		"echo 0 >> " + pfSysFsDevice + "/sriov_numvfs",
		"echo " + totalVfs + " >> " + pfSysFsDevice + "/sriov_numvfs",
	}

	for _, cmd := range cmds {
		if _, err := sshClient.Command(cmd, true); err != nil {
			return nil, err
		}
	}

	vfsString, err := sshClient.Command(`readlink -e `+pfSysFsDevice+`/virtfn*`, false)
	if err != nil {
		return nil, err
	}

	return strings.Split(vfsString, " "), nil
}

func (ks *KindSriov) bindToVfio(sshClient libssh.Client, sysFsDevice string) error {
	pciAddr, err := sshClient.Command("basename "+sysFsDevice, false)
	if err != nil {
		return err
	}

	driverPath := sysFsDevice + "/driver"
	driverOverride := sysFsDevice + "/driver_override"

	vfBusPciDeviceDriver, err := sshClient.Command("readlink "+driverPath+" | awk -F'/' '{print $NF}'", false)
	if err != nil {
		return err
	}
	vfBusPciDeviceDriver = strings.TrimSuffix(vfBusPciDeviceDriver, "\n")
	vfDriverName, err := sshClient.Command("basename "+vfBusPciDeviceDriver, false)
	if err != nil {
		return err
	}

	if _, err := sshClient.Command("[[ '"+vfDriverName+"' != 'vfio-pci' ]] && echo "+pciAddr+" > "+driverPath+"/unbind && echo 'vfio-pci' > "+driverOverride+" && echo "+pciAddr+" > /sys/bus/pci/drivers/vfio-pci/bind", true); err != nil {
		return err
	}

	return nil
}
