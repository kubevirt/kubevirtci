package utils

import (
	"strconv"

	"github.com/docker/go-connections/nat"
	"github.com/spf13/pflag"
)

// AppendIfExplicit append port to the portMap if the port flag exists
func AppendIfExplicit(ports nat.PortMap, exposedPort int, flagSet *pflag.FlagSet, flagName string) error {
	flag := flagSet.Lookup(flagName)
	if flag != nil && flag.Changed {
		publicPort, err := flagSet.GetUint(flagName)
		if err != nil {
			return err
		}
		port := TCPPortOrDie(exposedPort)
		ports[port] = []nat.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: strconv.Itoa(int(publicPort)),
			},
		}
	}
	return nil
}
