package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/coreos/go-iptables/iptables"
	"github.com/google/uuid"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"kubevirt.io/kubevirtci/cluster-provision/centos9/vmcli/cmd/qemu"
	"kubevirt.io/kubevirtci/cluster-provision/centos9/vmcli/cmd/utils"
)

const (
	// The path to the file that informs that the SSH script is ready
	sshReadyPath = "/ssh_ready"
	// The path to the script that connects to the guest via SSH
	sshScriptPath = "/usr/local/bin/ssh.sh"
	// The script that connects to the guest via SSH, formatted with the node number
	sshScriptContents = `#!/bin/bash
set -e
dockerize -wait tcp://192.168.66.1%02[1]d:22 -timeout 300s &>/dev/null
ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no vagrant@192.168.66.1%02[1]d -i vagrant.key -p 22 -q $@
`

	// The path to the legacy provisioned disk image
	provisionedDiskPath = "provisioned.qcow2"
	// The path to the first primary disk image
	firstDiskPath = "disk01.qcow2"
	// The default size of the primary disk
	defaultDiskSizeStr = "50Gi"

	// The path to the file that contains the kernel flags
	kernelArgsPath = "/kernel.args"
	// The path to the file that contains the additional kernel flags
	additionalKernelArgsPath = "/additional.kernel.args"
)

// NewRootCommand returns entrypoint command to interact with all other commands
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "vmcli",
		Short:         "vmcli creates a virtual machine that will host the k8s cluster",
		RunE:          run,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.Flags().StringP("memory", "m", "3096M", "amount of memory of the VM")
	root.Flags().Uint64P("cpu", "c", 2, "amount of CPU cores in the VM")
	root.Flags().Uint64P("numa", "a", 1, "amount of NUMA nodes")
	root.Flags().StringP("qemu-args", "q", "", "additional flags to pass to QEMU")
	root.Flags().StringP("additional-kernel-args", "k", "", "additional arguments passed to the kernel cmdline")
	root.Flags().StringP("next-disk", "n", "", "path to the primary disk image to create and attach to the VM")
	root.Flags().StringP("block-device", "b", "", "path to the block device image to create and attach to the VM")
	root.Flags().Uint64P("block-device-size", "s", 10737418240, "size of the block device image to create")
	root.Flags().UintSliceP("nvme-device-size", "e", []uint{}, "sizes of the NVMe device images to create")
	root.Flags().UintSliceP("scsi-device-size", "t", []uint{}, "sizes of the SCSI device images to create")
	root.Flags().UintSliceP("usb-device-size", "u", []uint{}, "sizes of the USB device images to create")

	return root

}

// Writes the SSH script to connect to the guest
func writeSshFiles(nodeNum int) error {
	if err := os.WriteFile(sshScriptPath, []byte(fmt.Sprintf(sshScriptContents, nodeNum)), 744); err != nil {
		return err
	}

	return os.WriteFile(sshReadyPath, []byte("done\n"), 0644)
}

// Add the iptables rule to forward SSH connections from the container to the guest
func addSshIptablesRule(iptables *iptables.IPTables, nodeNum int, rootless bool) error {
	err := iptables.AppendUnique("nat", "POSTROUTING", "!", "-s", "192.168.66.0/16", "--out-interface", "br0", "-j", "MASQUERADE")
	if err != nil {
		return err
	}

	dportFlag := fmt.Sprintf("22%02d", nodeNum)
	toDestinationFlag := fmt.Sprintf("192.168.66.1%02d:22", nodeNum)

	if rootless {
		// Add DNAT rule for rootless podman (traffic originating from loopback adapter)
		return iptables.AppendUnique("nat", "OUTPUT", "-p", "tcp", "--dport", dportFlag, "-j", "DNAT", "--to-destination", toDestinationFlag)
	}

	err = iptables.AppendUnique("filter", "FORWARD", "--in-interface", "eth0", "-j", "ACCEPT")
	if err != nil {
		return err
	}

	return iptables.AppendUnique("nat", "PREROUTING", "-p", "tcp", "-i", "eth0", "-m", "tcp", "--dport", dportFlag, "-j", "DNAT", "--to-destination", toDestinationFlag)
}

// Add the iptables rule to forward ports from the container to the guest
func addIptablesRules(iptables *iptables.IPTables, rootless bool, protocol string, ports []int) error {
	if rootless {
		for _, port := range ports {
			// Add DNAT rule for rootless podman (traffic originating from loopback adapter)
			toDestinationFlag := fmt.Sprintf("192.168.66.101:%d", port)

			err := iptables.AppendUnique("nat", "OUTPUT", "-p", protocol, "--dport", strconv.Itoa(port), "-j", "DNAT", "--to-destination", toDestinationFlag)
			if err != nil {
				return err
			}
		}

		return nil
	}

	for _, port := range ports {
		toDestinationFlag := fmt.Sprintf("192.168.66.101:%d", port)

		err := iptables.AppendUnique("nat", "PREROUTING", "-p", protocol, "-i", "eth0", "-m", protocol, "--dport", strconv.Itoa(port), "-j", "DNAT", "--to-destination", toDestinationFlag)
		if err != nil {
			return err
		}
	}

	return nil
}

func createPrimaryDisk(forcedNextDiskPath string) (string, error) {
	defaultDiskSizeQty, err := resource.ParseQuantity(defaultDiskSizeStr)
	if err != nil {
		return "", fmt.Errorf("Failed to parse the default primary disk size: %v", defaultDiskSizeStr)
	}

	defaultDiskSize := defaultDiskSizeQty.AsDec().UnscaledBig().Uint64()

	// For backward compatibility, so that we can just copy over the newer files
	_, err = os.Stat(provisionedDiskPath)

	if err == nil {
		os.Remove(firstDiskPath)
		os.Symlink(provisionedDiskPath, firstDiskPath)
	}

	diskUtils := utils.NewDiskUtil(afero.NewOsFs())
	diskFile, diskBackingFile, err := diskUtils.CalcNextDisk(".", forcedNextDiskPath)
	if err != nil {
		return "", fmt.Errorf("Failed to calculate the next disk image path: %v", err)
	}

	diskInfo, err := qemu.GetDiskInfo(diskBackingFile)
	if err != nil {
		return "", fmt.Errorf("Failed to get the disk image size of \"%s\": %v", diskBackingFile, err)
	}

	diskSize := diskInfo.VirtualSize
	if diskSize < defaultDiskSize {
		diskSize = defaultDiskSize
	}

	fmt.Printf("Creating disk \"%s backed by %s with size %d\"\n", diskFile, diskBackingFile, diskSize)
	return diskFile, qemu.CreateDiskWithBackingFile(diskFile, "qcow2", diskSize, diskBackingFile, "qcow2")
}

// Creates raw disks to be used by secondary block devices
func createSecondaryRawDisks(diskSizes []uint, deviceKind string) error {
	for i, size := range diskSizes {
		fmt.Printf("Creating disk \"%d\" for \"%s\" disk emulation\n", size, deviceKind)
		disk := fmt.Sprintf("/%s-%d.img", deviceKind, i)

		if err := qemu.CreateDisk(disk, "raw", uint64(size)); err != nil {
			return err
		}
	}

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	// Get the CLI flag values
	memoryFlag, err := cmd.Flags().GetString("memory")
	if err != nil {
		return err
	}

	cpuFlag, err := cmd.Flags().GetUint64("cpu")
	if err != nil {
		return err
	}

	numaFlag, err := cmd.Flags().GetUint64("numa")
	if err != nil {
		return err
	}

	qemuArgsFlag, err := cmd.Flags().GetString("qemu-args")
	if err != nil {
		return err
	}

	additionalKernelArgsFlag, err := cmd.Flags().GetString("additional-kernel-args")
	if err != nil {
		return err
	}

	nextDiskFlag, err := cmd.Flags().GetString("next-disk")
	if err != nil {
		return err
	}

	blockDeviceFlag, err := cmd.Flags().GetString("block-device")
	if err != nil {
		return err
	}

	blockDeviceSizeFlag, err := cmd.Flags().GetUint64("block-device-size")
	if err != nil {
		return err
	}

	nvmeDeviceSizeFlag, err := cmd.Flags().GetUintSlice("nvme-device-size")
	if err != nil {
		return err
	}

	scsiDeviceSizeFlag, err := cmd.Flags().GetUintSlice("scsi-device-size")
	if err != nil {
		return err
	}

	usbDeviceSizeFlag, err := cmd.Flags().GetUintSlice("usb-device-size")
	if err != nil {
		return err
	}

	envUtil := utils.NewEnvUtil(afero.NewOsFs())
	nodeNum, err := envUtil.GetNodeNb()
	if err != nil {
		return fmt.Errorf("Failed to get the number of this node: %v", err)
	}

	fmt.Println("Writing the SSH files")
	err = writeSshFiles(nodeNum)
	if err != nil {
		return fmt.Errorf("Failed to write the SSH files: %v", err)
	}

	err = envUtil.WaitForTap(nodeNum, -1)
	if err != nil {
		return fmt.Errorf("Failed to wait for the tap interface: %v", err)
	}

	rootless, err := envUtil.IsRootless()
	if err != nil {
		return fmt.Errorf("Failed to determine rootless status: %v", err)
	}
	fmt.Printf("Rootless mode: %t\n", rootless)

	iptables, err := iptables.New()
	if err != nil {
		return err
	}

	fmt.Println("Creating the common iptables rules")
	err = addSshIptablesRule(iptables, nodeNum, rootless)
	if err != nil {
		return err
	}

	// Route ports from container to VM for first node
	if nodeNum == 1 {
		fmt.Println("Creating the first node iptables rules")
		err = addIptablesRules(iptables, rootless, "tcp", []int{6443, 8443, 80, 443, 30007, 30008, 31001})
		if err != nil {
			return err
		}

		err = addIptablesRules(iptables, rootless, "udp", []int{31111})
		if err != nil {
			return err
		}
	}

	diskFile, err := createPrimaryDisk(nextDiskFlag)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("SSH will be available on container port 22%02d\n", nodeNum)
	fmt.Printf("VNC will be available on container port 59%02d\n", nodeNum)
	fmt.Printf("VM MAC in the guest network will be 52:55:00:d1:55:%02d\n", nodeNum)
	fmt.Printf("VM IP in the guest network will be 192.168.66.1%02d\n", nodeNum)
	fmt.Printf("VM hostname will be node%02d\n", nodeNum)
	fmt.Println()

	// Try to create /dev/kvm if it does not exist
	err = envUtil.EnsureKvmFileExists()
	if err != nil {
		return fmt.Errorf("Failed to ensure that the KVM file exists: %v", err)
	}

	// Prevent the emulated soundcard from messing with host sound
	os.Setenv("QEMU_AUDIO_DRV", "none")

	// Get the kernel args from the files
	kernelArgs, err := os.ReadFile(kernelArgsPath)
	if err != nil {
		return fmt.Errorf("Failed to read the kernel args from the file \"%s\": %v", kernelArgsPath, err)
	}

	additionalKernelArgs, err := os.ReadFile(additionalKernelArgsPath)
	if err != nil {
		return fmt.Errorf("Failed to read the additional kernel args from the file \"%s\": %v", additionalKernelArgsPath, err)
	}

	// Collect the QEMU command line
	systemUuid, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("Failed to create the random system UUID: %v", err)
	}

	qemuSystem := qemu.QemuSystem{
		Arch:          "x86_64",
		Memory:        memoryFlag,
		CpuCount:      cpuFlag,
		Numa:          numaFlag,
		KvmEnabled:    true,
		CpuModel:      "host,migratable=no,+invtsc",
		Machine:       "q35,accel=kvm,kernel_irqchip=split",
		SystemUuid:    systemUuid,
		VncServer:     fmt.Sprintf(":%02d", nodeNum),
		SerialHostDev: "pty",
		InitrdPath:    "/initrd.img",
		KernelPath:    "/vmlinuz",
		KernelArgs: []string{
			string(kernelArgs),
			string(additionalKernelArgs),
			additionalKernelArgsFlag,
		},
		Drives: []string{
			fmt.Sprintf("format=qcow2,file=%s,if=virtio,cache=unsafe", diskFile),
		},
		Devices: []string{
			fmt.Sprintf("virtio-net-pci,netdev=network0,mac=52:55:00:d1:55:%02d", nodeNum),
			"virtio-rng-pci",
			"intel-iommu,intremap=on,caching-mode=on",
			"intel-hda",
			"hda-duplex",
			"AC97",
		},
		Netdev: fmt.Sprintf("tap,id=network0,ifname=tap%02d,script=no,downscript=no", nodeNum),
	}

	// Create the secondary disks
	if blockDeviceFlag != "" {
		blockDeviceSize := uint64(10737418240) // 10Gi default
		if blockDeviceSizeFlag != 0 {
			blockDeviceSize = blockDeviceSizeFlag
		}

		fmt.Printf("Creating secondary disk \"%s with size %d\"\n", blockDeviceFlag, blockDeviceSize)
		err = qemu.CreateDisk(blockDeviceFlag, "qcow2", blockDeviceSize)
		if err != nil {
			return err
		}

		qemuSystem.Drives = append(qemuSystem.Drives, fmt.Sprintf("format=qcow2,file=%s,if=virtio,cache=unsafe", blockDeviceFlag))
	}

	createSecondaryRawDisks(nvmeDeviceSizeFlag, "nvme")
	createSecondaryRawDisks(scsiDeviceSizeFlag, "scsi")
	createSecondaryRawDisks(usbDeviceSizeFlag, "usb")

	// Start QEMU
	generatedQemuCmdline, err := qemuSystem.GenerateCmdline()
	if err != nil {
		return err
	}

	qemuCmdline := fmt.Sprintf("%s %s", generatedQemuCmdline, qemuArgsFlag)

	fmt.Println("Starting QEMU with the following arguments:")
	fmt.Println(qemuCmdline)
	fmt.Println()

	qemuCmd := exec.Command("sh", "-c", fmt.Sprintf("exec %s", qemuCmdline))
	qemuCmd.Stdin = os.Stdin
	qemuCmd.Stdout = os.Stdout
	qemuCmd.Stderr = os.Stderr
	return qemuCmd.Run()
}

// Execute executes root command
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
