package containers

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/net/context"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
)

type forwardDestination struct {
	Port string
	Host string
}

type portForward map[string]forwardDestination

type DNSMasqOptions struct {
	ClusterImage       string
	NodeCount          uint
	SecondaryNicsCount uint
	RandomPorts        bool
	PortMap            nat.PortMap
	PortForward        []string
	Prefix             string
}

func DNSMasq(cli *client.Client, ctx context.Context, options *DNSMasqOptions) (*container.ContainerCreateCreatedBody, error) {
	// Mount /lib/modules at dnsmasq if it's there since sometimes
	// some kernel modules may be mounted
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

	// Start dnsmasq
	config := &container.Config{
		Image: options.ClusterImage,
		Env: []string{
			fmt.Sprintf("NUM_NODES=%d", options.NodeCount),
			fmt.Sprintf("NUM_SECONDARY_NICS=%d", options.SecondaryNicsCount),
			fmt.Sprintf("PORT_FORWARD=%s", options.PortForward),
		},
		Cmd: []string{"/bin/bash", "-c", "/dnsmasq.sh"},
		ExposedPorts: nat.PortSet{
			utils.TCPPortOrDie(utils.PortSSH):        {},
			utils.TCPPortOrDie(utils.PortRegistry):   {},
			utils.TCPPortOrDie(utils.PortOCP):        {},
			utils.TCPPortOrDie(utils.PortAPI):        {},
			utils.TCPPortOrDie(utils.PortVNC):        {},
			utils.TCPPortOrDie(utils.PortHTTP):       {},
			utils.TCPPortOrDie(utils.PortHTTPS):      {},
			utils.TCPPortOrDie(utils.PortPrometheus): {},
			utils.TCPPortOrDie(utils.PortGrafana):    {}, utils.TCPPortOrDie(utils.PortUploadProxy): {},
		},
	}

	hostConfig := &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: options.RandomPorts,
		PortBindings:    options.PortMap,
		ExtraHosts: []string{
			"nfs:192.168.66.2",
			"registry:192.168.66.2",
			"ceph:192.168.66.2",
		},
		Mounts: dnsmasqMounts,
	}

	portForward, err := parsePortForward(options.PortForward)
	if err != nil {
		return nil, err
	}

	for p, d := range portForward {
		exposedPort, err := nat.NewPort("tcp", p)
		if err != nil {
			return nil, err
		}
		config.ExposedPorts[exposedPort] = struct{}{}
		hostConfig.PortBindings[exposedPort] = []nat.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: d.Port,
			},
		}

	}
	dnsmasq, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, options.Prefix+"-dnsmasq")
	if err != nil {
		return nil, err
	}
	return &dnsmasq, nil
}

func parsePortForward(portForwardSerialized []string) (portForward, error) {
	pf := portForward{}
	for _, pfs := range portForwardSerialized {
		splitted := strings.SplitN(pfs, ":", 2)
		if len(splitted) != 2 {
			continue
		}
		fmt.Println(splitted)
		containerPort := splitted[0]
		destinationHostPort := splitted[1]
		fmt.Println(destinationHostPort)
		host, port, err := net.SplitHostPort(destinationHostPort)
		if err != nil {
			return pf, fmt.Errorf("failed parsing port forward: %v", err)
		}
		pf[containerPort] = forwardDestination{
			Port: port,
			Host: host,
		}
	}
	return pf, nil
}
