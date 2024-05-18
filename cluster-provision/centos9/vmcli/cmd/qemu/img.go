package qemu

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Wrapper around the output of the "qemu-img info" command
type qemuImgInfo struct {
	// The virtual size of the disk image
	VirtualSize uint64 `json:"virtual-size"`
}

// Creates a virtual disk image
func CreateDisk(path string, format string, size uint64) error {
	cmd := exec.Command("qemu-img", "create", "-f", format, path, strconv.FormatUint(size, 10))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Creates a virtual disk image that is backed by another disk image
func CreateDiskWithBackingFile(path string, format string, size uint64, backingPath string, backingFormat string) error {
	cmd := exec.Command("qemu-img", "create", "-f", format, "-o", fmt.Sprintf("backing_file=%s", backingPath), "-F", backingFormat, path, strconv.FormatUint(size, 10))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Parse a JSON qemu-img info object
func ParseDiskInfo(b []byte) (qemuImgInfo, error) {
	var info qemuImgInfo
	err := json.Unmarshal(b, &info)

	return info, err
}

// Get information about the specified disk image on the file system
func GetDiskInfo(diskPath string) (qemuImgInfo, error) {
	cmd := exec.Command("qemu-img", "info", "--output", "json", diskPath)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		return qemuImgInfo{}, err
	}

	return ParseDiskInfo(out)
}
