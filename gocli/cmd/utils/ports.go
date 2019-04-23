package utils

import (
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
)

const (
	// PortSSH contains SSH port
	PortSSH = 2201
	// PortRegistry contains SSH port
	PortRegistry = 5000
	// PortOCP contains SSH port
	PortOCP = 8443
	// PortAPI contains SSH port
	PortAPI = 6443
	// PortVNC contains SSH port
	PortVNC = 5901

	// PortNameSSH contains SSH port name
	PortNameSSH = "ssh"
	// PortNameOCP contains OCP port name
	PortNameOCP = "ocp"
	// PortNameRegistry contains registry port name
	PortNameRegistry = "registry"
	// PortNameAPI contains API port name
	PortNameAPI = "api"
	// PortNameVNC contains VNC port name
	PortNameVNC = "vnc"
)

// GetPublicPort returns public port by private port
func GetPublicPort(port uint16, ports []types.Port) (uint16, error) {
	for _, p := range ports {
		if p.PrivatePort == port {
			return p.PublicPort, nil
		}
	}
	return 0, fmt.Errorf("port is not exposed")
}

// PrintPublicPort prints public port
func PrintPublicPort(port uint16, ports []types.Port) error {
	p, err := GetPublicPort(port, ports)
	if err != nil {
		return err
	}
	fmt.Println(p)
	return nil
}

// TCPPortOrDie returns net.Port object or panic if cast failed
func TCPPortOrDie(port int) nat.Port {
	p, err := nat.NewPort("tcp", strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	return p
}
