package utils

import (
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
)

const (
	// PortSSH contains SSH port for the master node
	PortSSH = 2201
	// PortSSHWorker contains SSH port for the worker node
	PortSSHWorker = 2202
	// PortRegistry contains private image registry port
	PortRegistry = 5000
	// PortOCP contains OCP API server port
	PortOCP = 8443
	// PortAPI contains API server port
	PortAPI = 6443
	// PortVNC contains first VM VNC port
	PortVNC = 5901
	// PortHTTP contains ingress HTTP port
	PortHTTP = 80
	// PortHTTPS contains ingress HTTPS port
	PortHTTPS = 443
	//PortOCPConsole contains OCP console port
	PortOCPConsole = 443

	// PortNameSSH contains master node SSH port name
	PortNameSSH = "ssh"
	// PortNameSSHWorker contains worker node SSH port name
	PortNameSSHWorker = "ssh-worker"
	// PortNameOCP contains OCP port name
	PortNameOCP = "ocp"
	// PortNameRegistry contains registry port name
	PortNameRegistry = "registry"
	// PortNameAPI contains API port name
	// TODO: change the name to API
	PortNameAPI = "k8s"
	// PortNameVNC contains VNC port name
	PortNameVNC = "vnc"
	// PortNameHTTP contains HTTP port name
	PortNameHTTP = "http"
	// PortNameHTTPS contains HTTPS port name
	PortNameHTTPS = "https"
	// PortNameOCPConsole contains OCP console port
	PortNameOCPConsole = "console"
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
