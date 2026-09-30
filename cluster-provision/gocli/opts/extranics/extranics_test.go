package extranics

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		secondaryNics uint
		wantErr       string
	}{
		{
			name:          "empty config with zero nics",
			secondaryNics: 0,
		},
		{
			name: "ifaces covered by secondary nics",
			config: Config{
				BridgedIfaces:   []string{"eth1"},
				AddressedIfaces: []string{"eth2", "eth3"},
			},
			secondaryNics: 3,
		},
		{
			name: "unused secondary nics are fine",
			config: Config{
				BridgedIfaces: []string{"eth1"},
			},
			secondaryNics: 3,
		},
		{
			name: "iface beyond secondary nics count",
			config: Config{
				BridgedIfaces:   []string{"eth1"},
				AddressedIfaces: []string{"eth2", "eth3"},
			},
			secondaryNics: 0,
			wantErr:       "eth1 requires --secondary-nics >= 1 (got 0); set KUBEVIRT_NUM_SECONDARY_NICS accordingly",
		},
		{
			name: "highest addressed iface beyond count",
			config: Config{
				AddressedIfaces: []string{"eth3"},
			},
			secondaryNics: 2,
			wantErr:       "eth3 requires --secondary-nics >= 3 (got 2); set KUBEVIRT_NUM_SECONDARY_NICS accordingly",
		},
		{
			name: "invalid iface name",
			config: Config{
				BridgedIfaces: []string{"eth0"},
			},
			secondaryNics: 1,
			wantErr:       `"eth0" is not a secondary interface name`,
		},
		{
			name: "iface both bridged and addressed",
			config: Config{
				BridgedIfaces:   []string{"eth1"},
				AddressedIfaces: []string{"eth1"},
			},
			secondaryNics: 1,
			wantErr:       `"eth1" cannot be both bridged and addressed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config, tt.secondaryNics)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got nil", tt.wantErr)
			}
			if err.Error() != tt.wantErr {
				t.Fatalf("expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
