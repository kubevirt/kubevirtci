// Package extranics renders the script that configures the secondary network
// interfaces of a node. It is shared by the node01 and the nodes provisioners,
// which configure their nodes alike.
package extranics

import (
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"text/template"
)

//go:embed scripts/configure-extra-nics.sh.tmpl
var scriptTemplate string

// The interface names end up interpolated into a shell script, so accept only
// the shape the provider hands out and reject anything else.
var ifaceNameRegex = regexp.MustCompile(`^eth[1-9][0-9]*$`)

// Config lists what is to be done with each of the secondary interfaces.
type Config struct {
	// BridgedIfaces are enslaved to a bridge of their own, brX for ethX.
	BridgedIfaces []string

	// AddressedIfaces get an address of their own, derived from eth0.
	AddressedIfaces []string
}

// Empty tells whether there is any interface to configure.
func (c Config) Empty() bool {
	return len(c.BridgedIfaces) == 0 && len(c.AddressedIfaces) == 0
}

// Script renders the configuration script for the given interfaces.
func Script(config Config) (string, error) {
	if err := validate(config); err != nil {
		return "", err
	}

	tmpl, err := template.New("configure-extra-nics").Parse(scriptTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing the extra NICs script template: %v", err)
	}

	script := &bytes.Buffer{}
	if err := tmpl.Execute(script, config); err != nil {
		return "", fmt.Errorf("rendering the extra NICs script: %v", err)
	}

	return script.String(), nil
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
