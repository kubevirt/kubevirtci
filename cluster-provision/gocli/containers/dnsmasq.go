package containers

import (
	"fmt"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/net/context"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
)

type DNSMasqOptions struct {
	ClusterImage       string
	NodeCount          uint
	SecondaryNicsCount uint
	RandomPorts        bool
	PortMap            nat.PortMap
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
	dnsmasq, err := cli.ContainerCreate(ctx, &container.Config{
		Image: options.ClusterImage,
		Env: []string{
			fmt.Sprintf("NUM_NODES=%d", options.NodeCount),
			fmt.Sprintf("NUM_SECONDARY_NICS=%d", options.SecondaryNicsCount),
		},
		Cmd: []string{"/bin/bash", "-c", "/dnsmasq.sh"},
		ExposedPorts: nat.PortSet{
			utils.TCPPortOrDie(utils.PortSSH):      {},
			utils.TCPPortOrDie(utils.PortRegistry): {},
			utils.TCPPortOrDie(utils.PortOCP):      {},
			utils.TCPPortOrDie(utils.PortAPI):      {},
			utils.TCPPortOrDie(utils.PortVNC):      {},
			utils.TCPPortOrDie(utils.PortHTTP):     {},
			utils.TCPPortOrDie(utils.PortHTTPS):    {},
		},
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: options.RandomPorts,
		PortBindings:    options.PortMap,
		ExtraHosts: []string{
			"nfs:192.168.66.2",
			"registry:192.168.66.2",
			"ceph:192.168.66.2",
		},
		Mounts: dnsmasqMounts,
	}, nil, options.Prefix+"-dnsmasq")
	if err != nil {
		return nil, err
	}
	return &dnsmasq, nil
}
