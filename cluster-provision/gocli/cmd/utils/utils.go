package utils

import (
	"fmt"
	"os"
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

// ForwardEnv returns KEY=VALUE strings for env vars that are set in the current process.
func ForwardEnv(names ...string) []string {
	var result []string
	for _, name := range names {
		if val, ok := os.LookupEnv(name); ok {
			result = append(result, fmt.Sprintf("%s=%s", name, val))
		}
	}
	return result
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
