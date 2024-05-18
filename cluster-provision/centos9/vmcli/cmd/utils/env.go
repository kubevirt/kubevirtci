package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/sys/unix"
)

const (
	// Environment variable key of the node number
	NodeNumKey = "NODE_NUM"

	// Path where to check whether a network interface is present
	NetPath = "/sys/class/net/"
	// Name of the tap interfaces, formatted with the number of the tap interface
	TapName = "tap%02d"

	// Path to the .containerenv file
	ContainerenvPath = "/run/.containerenv"

	// Path to the KVM file
	kvmPath = "/dev/kvm"
)

// Util methods for the environment
type EnvUtil struct {
	fs afero.Fs
}

// Creates a new struct for environment utils
func NewEnvUtil(fs afero.Fs) *EnvUtil {
	return &EnvUtil{
		fs,
	}
}

// Returns the number of this node
func (eu EnvUtil) GetNodeNb() (int, error) {
	nodeNumStr := os.Getenv(NodeNumKey)

	if nodeNumStr != "" {
		nodeNum, err := strconv.Atoi(nodeNumStr)
		if err != nil {
			return 0, err
		}

		return nodeNum, nil
	}

	return 1, nil
}

// Wait for a tap network interface to be present
func (eu EnvUtil) WaitForTap(tapNum int, maxTries int) error {
	// Wait for the correct /sys/class/net/ directory to exist
	tapPath := filepath.Join(NetPath, fmt.Sprintf(TapName, tapNum))
	tries := maxTries
	_, err := eu.fs.Stat(tapPath)

	for tries > 0 && errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Waiting for %s to become ready\n", fmt.Sprintf(TapName, tapNum))
		time.Sleep(time.Duration(100) * time.Millisecond)

		_, err = eu.fs.Stat(tapPath)
		tries -= 1
	}

	if !errors.Is(err, os.ErrNotExist) || tries == 0 {
		return err
	}

	return nil
}

// Returns whether the container is running in a rootless environment
func (eu EnvUtil) IsRootless() (bool, error) {
	f, err := eu.fs.Open(ContainerenvPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// If the file doesn't exist, assume we're not running in rootless mode
			return false, nil
		} else {
			return false, err
		}
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		value, found := strings.CutPrefix(line, "rootless=")

		if found {
			return value == "1", nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	// Didn't find the rootless line, assume we're not running in rootless mode
	return false, nil
}

// Ensures that the KVM file exists
func (eu EnvUtil) EnsureKvmFileExists() error {
	_, err := eu.fs.Stat(kvmPath)

	if err == nil || !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return unix.Mknod(kvmPath, syscall.S_IFCHR|0666, int(unix.Mkdev(10, 232)))
}
