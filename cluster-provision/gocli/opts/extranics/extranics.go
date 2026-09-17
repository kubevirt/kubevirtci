// Package extranics renders the script that configures the secondary network
// interfaces of a node. It is shared by the node01 and the nodes provisioners,
// which configure their nodes alike.
package extranics

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"
)

// The third octet of the subnet dnsmasq leases eth0 from. The secondary
// subnets follow it, one per interface index.
const primarySubnetOctet = 66

//go:embed scripts/configure-extra-nics.sh.tmpl
var scriptTemplate string

// The interface names end up interpolated into a shell script, so accept only
// the shape the provider hands out and reject anything else.
var ifaceNameRegex = regexp.MustCompile(`^eth[1-9][0-9]*$`)

// Config lists what is to be done with each of the secondary interfaces.
type Config struct {
	// NodeIdx is the index of the node the script is rendered for. It decides
	// the host part of the addresses, as it does for the eth0 lease.
	NodeIdx int

	// BridgedIfaces are enslaved to a bridge of their own, brX for ethX.
	BridgedIfaces []string

	// AddressedIfaces get an address of their own instead of a bridge.
	AddressedIfaces []string
}

type addressedIface struct {
	Name string
	IPv4 string
	IPv6 string
}

type templateData struct {
	BridgedIfaces   []string
	AddressedIfaces []addressedIface
}

// Empty tells whether there is any interface to configure.
func (c Config) Empty() bool {
	return len(c.BridgedIfaces) == 0 && len(c.AddressedIfaces) == 0
}

// Validate checks that the interfaces to configure are well formed and that
// each of them is covered by the secondary NICs the VMs are started with.
// Call this once flags are parsed so a mismatch fails before any node boots.
func Validate(config Config, secondaryNics uint) error {
	if err := validate(config); err != nil {
		return err
	}

	for _, iface := range slices.Concat(config.BridgedIfaces, config.AddressedIfaces) {
		idx, err := ifaceIndex(iface)
		if err != nil {
			return err
		}
		if uint(idx) > secondaryNics {
			return fmt.Errorf("%s requires --secondary-nics >= %d (got %d); set KUBEVIRT_NUM_SECONDARY_NICS accordingly",
				iface, idx, secondaryNics)
		}
	}

	return nil
}

// Script renders the configuration script for the given interfaces.
func Script(config Config) (string, error) {
	if err := validate(config); err != nil {
		return "", err
	}

	data := templateData{BridgedIfaces: config.BridgedIfaces}
	for _, iface := range config.AddressedIfaces {
		ipv4, ipv6, err := addresses(config.NodeIdx, iface)
		if err != nil {
			return "", err
		}
		data.AddressedIfaces = append(data.AddressedIfaces, addressedIface{Name: iface, IPv4: ipv4, IPv6: ipv6})
	}

	tmpl, err := template.New("configure-extra-nics").Parse(scriptTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing the extra NICs script template: %v", err)
	}

	script := &bytes.Buffer{}
	if err := tmpl.Execute(script, data); err != nil {
		return "", fmt.Errorf("rendering the extra NICs script: %v", err)
	}

	return script.String(), nil
}

// addresses of a secondary interface, spelled out the way dnsmasq spells out
// the eth0 lease of every node: 192.168.66.1<nn> and fd00::1<nn> for node
// <nn>. Each interface sits one subnet further, so eth2 of node01 holds
// 192.168.68.101 and fd00:2::101.
func addresses(nodeIdx int, iface string) (string, string, error) {
	ifaceIdx, err := ifaceIndex(iface)
	if err != nil {
		return "", "", err
	}

	subnetOctet := primarySubnetOctet + ifaceIdx
	if subnetOctet > 255 {
		return "", "", fmt.Errorf("%q falls past the last subnet, 192.168.255.0/24", iface)
	}

	host := fmt.Sprintf("1%02d", nodeIdx)

	return fmt.Sprintf("192.168.%d.%s/24", subnetOctet, host),
		fmt.Sprintf("fd00:%d::%s/64", ifaceIdx, host),
		nil
}

func ifaceIndex(iface string) (int, error) {
	idx, err := strconv.Atoi(strings.TrimPrefix(iface, "eth"))
	if err != nil {
		return 0, fmt.Errorf("no interface index in %q: %v", iface, err)
	}
	return idx, nil
}

func validate(config Config) error {
	bridged := map[string]bool{}
	for _, iface := range config.BridgedIfaces {
		if !ifaceNameRegex.MatchString(iface) {
			return fmt.Errorf("%q is not a secondary interface name", iface)
		}
		bridged[iface] = true
	}

	for _, iface := range config.AddressedIfaces {
		if !ifaceNameRegex.MatchString(iface) {
			return fmt.Errorf("%q is not a secondary interface name", iface)
		}
		if bridged[iface] {
			return fmt.Errorf("%q cannot be both bridged and addressed", iface)
		}
	}

	return nil
}
