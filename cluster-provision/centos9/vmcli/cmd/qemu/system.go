package qemu

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// The name of the QEMU main executable
// TODO: On RHEL and similar, the executable is "qemu-kvm". We'll need to check the distribution and use the correct executable.
const qemuExec = "qemu-system-%s"

// Thin wrapper around the "qemu-system-x" command
type QemuSystem struct {
	// The architecture of the "qemu-system" command to start
	Arch string
	// The memory to allocate to the guest, wraps the "-m" argument
	Memory string
	// The number of the CPUs in the SMP system, wraps the "-smp" argument
	CpuCount uint64
	// The number of NUMA nodes in the system, wraps "-object memory-backend-ram" and "-numa node" arguments
	Numa uint64
	// Whether to use KVM, wraps the "-enable-kvm" argument
	KvmEnabled bool
	// Select the CPU model, wraps the "-cpu" argument
	CpuModel string
	// Select the emulated machine, wraps the "-machine" argument
	Machine string
	// Set the system UUID, wraps the "-uuid" argument
	SystemUuid uuid.UUID
	// Start a VNC server on the specified display, wraps the "-vnc" argument
	VncServer string
	// Redirect the virtual serial port to the specified character device, wraps the "-serial" argument
	SerialHostDev string
	// Path to the initramfs file, wraps the "-initrd" argument
	InitrdPath string
	// Path to the kernel to boot, wraps the "-kernel" argument
	KernelPath string
	// The cmdlines to pass to the booted kernel, wraps the "-append" argument
	KernelArgs []string
	// Drives to attach to the guest, wraps the "-drive" arguments
	Drives []string
	// Devices to attach to the guest, wraps the "-device" arguments
	Devices []string
	// The network backend to configure, wraps the "-netdev" argument
	Netdev string
}

// Parse the memory argument into a value and possibly a unit
func (q QemuSystem) ParseMemory() (uint64, string, error) {
	regex, err := regexp.Compile(`(\d+)(\w*)`)
	if err != nil {
		return 0, "", err
	}

	submatch := regex.FindStringSubmatch(q.Memory)
	if len(submatch) < 2 {
		return 0, "", fmt.Errorf("Unable to parse the QEMU memory argument %q", q.Memory)
	}

	val, err := strconv.ParseUint(submatch[1], 10, 64)
	if err != nil {
		return 0, "", err
	}

	unit := ""
	if len(submatch) >= 3 {
		unit = submatch[2]
	}

	return val, unit, nil
}

// Generate the QEMU arguments for creating the NUMA topology
func (q QemuSystem) GenerateNumaArguments() ([]string, error) {
	if q.Numa < 2 {
		return []string{}, nil
	}

	result := []string{}

	memoryValue, memoryUnit, err := q.ParseMemory()
	if err != nil {
		return []string{}, err
	}

	if q.CpuCount%q.Numa > 0 || memoryValue%q.Numa > 0 {
		return []string{}, fmt.Errorf("unable to calculate symmetric NUMA topology with vCPUs:%v Memory:%v NUMA:%v", q.CpuCount, q.Memory, q.Numa)
	}

	memoryPerNodeValue := memoryValue / q.Numa
	memoryPerNode := fmt.Sprintf("%v%v", memoryPerNodeValue, memoryUnit)
	cpuPerNode := q.CpuCount / q.Numa

	for nodeId := uint64(0); nodeId < q.Numa; nodeId++ {
		nodeFirstCpu := nodeId * cpuPerNode
		nodeLastCpu := nodeFirstCpu + cpuPerNode - 1

		memoryBackendRamArg := fmt.Sprintf("-object memory-backend-ram,size=%v,id=m%v", memoryPerNode, nodeId)
		numaNodeArg := fmt.Sprintf("-numa node,nodeid=%v,memdev=m%v,cpus=%v-%v", nodeId, nodeId, nodeFirstCpu, nodeLastCpu)

		result = append(result, memoryBackendRamArg, numaNodeArg)
	}

	return result, nil
}

// Generate the command line to use to start QEMU
func (q QemuSystem) GenerateCmdline() (string, error) {
	qemuArgs := []string{
		fmt.Sprintf(qemuExec, q.Arch),
		"-m", q.Memory,
		"-smp", strconv.FormatUint(q.CpuCount, 10),
		"-cpu", q.CpuModel,
		"-M", q.Machine,
		"-uuid", q.SystemUuid.String(),
		"-vnc", q.VncServer,
		"-serial", q.SerialHostDev,
		"-initrd", q.InitrdPath,
		"-kernel", q.KernelPath,
		"-append", fmt.Sprintf("\"%s\"", strings.TrimSpace(strings.Join(q.KernelArgs, " "))),
		"-netdev", q.Netdev,
	}

	if q.KvmEnabled {
		qemuArgs = append(qemuArgs, "-enable-kvm")
	}

	for _, drive := range q.Drives {
		qemuArgs = append(qemuArgs, "-drive", drive)
	}

	for _, device := range q.Devices {
		qemuArgs = append(qemuArgs, "-device", device)
	}

	numaArgs, err := q.GenerateNumaArguments()
	if err != nil {
		return "", err
	}

	if len(numaArgs) > 0 {
		qemuArgs = append(qemuArgs, numaArgs...)
	}

	return strings.Join(qemuArgs, " "), nil
}
