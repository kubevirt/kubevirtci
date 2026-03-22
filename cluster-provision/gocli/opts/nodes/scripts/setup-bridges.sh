#!/bin/bash
set -e

# Setup bridges for secondary network interfaces (eth1, eth2, eth3, ...)
# For each ethX interface where X >= 1, create a corresponding brX bridge
# Uses NetworkManager (nmcli) to create persistent bridge configurations

for iface in $(ls /sys/class/net/ 2>/dev/null | grep -E '^eth[1-9]$|^eth[0-9]{2,}$'); do
    bridge_name="br${iface#eth}"

    echo "Setting up bridge ${bridge_name} for interface ${iface}"

    # Delete existing connections for the interface if any (to ensure clean state)
    nmcli connection delete ${iface} 2>/dev/null || true
    nmcli connection delete ${bridge_name} 2>/dev/null || true

    # Create bridge connection (disable DHCP to avoid timeouts)
    nmcli connection add type bridge ifname ${bridge_name} con-name ${bridge_name} ipv4.method disabled ipv6.method disabled

    # Add ethernet interface as bridge slave
    nmcli connection add type ethernet ifname ${iface} con-name ${iface} master ${bridge_name}

    # Bring up the bridge (which will also bring up the slave)
    nmcli connection up ${bridge_name}

    echo "Bridge ${bridge_name} created and configured successfully"
done

echo "All secondary interface bridges configured"
