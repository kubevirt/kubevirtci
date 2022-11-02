package utils

import (
	"strconv"

	"github.com/docker/go-connections/nat"
	"github.com/spf13/pflag"
)

// AppendTCPIfExplicit append TCP port to the portMap if the port flag exists
func AppendTCPIfExplicit(ports nat.PortMap, exposedPort int, flagSet *pflag.FlagSet, flagName string) error {
	return appendIfExplicit(ports, exposedPort, flagSet, flagName, TCPPortOrDie)
}

// AppendUDPIfExplicit append UDP port to the portMap if the port flag exists
func AppendUDPIfExplicit(ports nat.PortMap, exposedPort int, flagSet *pflag.FlagSet, flagName string) error {
	return appendIfExplicit(ports, exposedPort, flagSet, flagName, UDPPortOrDie)
}

func appendIfExplicit(ports nat.PortMap, exposedPort int, flagSet *pflag.FlagSet, flagName string, portFn func(port int) nat.Port) error {
	flag := flagSet.Lookup(flagName)
	if flag != nil && flag.Changed {
		publicPort, err := flagSet.GetUint(flagName)
		if err != nil {
			return err
		}
		port := portFn(exposedPort)
		ports[port] = []nat.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: strconv.Itoa(int(publicPort)),
			},
		}
	}
	return nil
}
